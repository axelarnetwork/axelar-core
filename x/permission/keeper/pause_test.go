package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
	"github.com/axelarnetwork/axelar-core/x/permission/types/mock"
)

func setupPauseTest(t *testing.T) (sdk.Context, keeper.Keeper, types.MsgServer, *mock.UpgradeKeeperMock) {
	encCfg := app.MakeEncodingConfig()
	key := store.NewKVStoreKey("permission")
	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, key, store.NewKVStoreKey("tpermissionKey"), "permission")
	k := keeper.NewKeeper(0, encCfg.Codec, key, store.NewTransientStoreKey(types.TStoreKey), subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), cmtproto.Header{Height: rand.I64Between(1, 1_000_000)}, false, log.NewTestLogger(t))

	upgradeK := &mock.UpgradeKeeperMock{
		GetUpgradePlanFunc: func(context.Context) (upgradetypes.Plan, error) {
			return upgradetypes.Plan{}, upgradetypes.ErrNoUpgradePlanFound
		},
		ScheduleUpgradeFunc: func(context.Context, upgradetypes.Plan) error { return nil },
	}

	return ctx, k, keeper.NewMsgServerImpl(k, upgradeK), upgradeK
}

func findEvent(ctx sdk.Context, eventType string) (map[string]string, bool) {
	for _, event := range ctx.EventManager().Events() {
		if event.Type != eventType {
			continue
		}

		attributes := map[string]string{}
		for _, attr := range event.Attributes {
			attributes[attr.Key] = attr.Value
		}
		return attributes, true
	}

	return nil, false
}

func TestMsgServer_PauseAndUnpause(t *testing.T) {
	ctx, k, server, _ := setupPauseTest(t)
	sender := rand.AccAddr().String()

	assert.Equal(t, exported.PAUSE_LEVEL_UNSPECIFIED, k.GetPauseLevel(ctx))

	_, err := server.Unpause(ctx, &types.UnpauseRequest{Sender: sender, Reason: "nothing to unpause"})
	assert.ErrorContains(t, err, "not paused")

	_, err = server.Pause(ctx, &types.PauseRequest{Sender: sender, Level: exported.PAUSE_LEVEL_ALL, Reason: "incident"})
	require.NoError(t, err)
	assert.Equal(t, exported.PAUSE_LEVEL_ALL, k.GetPauseLevel(ctx))

	event, ok := findEvent(ctx, types.EventTypeEmergencyPause)
	require.True(t, ok)
	assert.Equal(t, sender, event[types.AttributeKeySigner])
	assert.Equal(t, "incident", event[types.AttributeKeyReason])
	assert.Equal(t, exported.PAUSE_LEVEL_ALL.String(), event[types.AttributeKeyPauseLevel])
	assert.Equal(t, exported.PAUSE_LEVEL_UNSPECIFIED.String(), event[types.AttributeKeyPreviousPauseLevel])

	_, err = server.Pause(ctx, &types.PauseRequest{Sender: sender, Level: exported.PAUSE_LEVEL_ALL, Reason: "again"})
	assert.ErrorContains(t, err, "already paused")

	_, err = server.Pause(ctx, &types.PauseRequest{Sender: sender, Level: exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, Reason: "loosen"})
	require.NoError(t, err)
	assert.Equal(t, exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, k.GetPauseLevel(ctx))

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	_, err = server.Unpause(ctx, &types.UnpauseRequest{Sender: sender, Reason: "resolved"})
	require.NoError(t, err)
	assert.Equal(t, exported.PAUSE_LEVEL_UNSPECIFIED, k.GetPauseLevel(ctx))

	event, ok = findEvent(ctx, types.EventTypeEmergencyUnpause)
	require.True(t, ok)
	assert.Equal(t, sender, event[types.AttributeKeySigner])
	assert.Equal(t, "resolved", event[types.AttributeKeyReason])
	assert.Equal(t, exported.PAUSE_LEVEL_EXCEPT_VALIDATORS.String(), event[types.AttributeKeyPreviousPauseLevel])
}

func TestKeeper_GetPauseLevel_ConsumesNoGas(t *testing.T) {
	ctx, k, server, _ := setupPauseTest(t)

	_, err := server.Pause(ctx, &types.PauseRequest{Sender: rand.AccAddr().String(), Level: exported.PAUSE_LEVEL_ALL, Reason: "incident"})
	require.NoError(t, err)

	ctx = ctx.WithGasMeter(store.NewGasMeter(1_000_000))
	assert.Equal(t, exported.PAUSE_LEVEL_ALL, k.GetPauseLevel(ctx))
	assert.Zero(t, ctx.GasMeter().GasConsumed())
}

func TestMsgServer_HaltChain(t *testing.T) {
	t.Run("schedules a plan at the next height", func(t *testing.T) {
		ctx, _, server, upgradeK := setupPauseTest(t)
		sender := rand.AccAddr().String()

		res, err := server.HaltChain(ctx, &types.HaltChainRequest{Sender: sender, Reason: "incident"})
		require.NoError(t, err)

		haltHeight := ctx.BlockHeight() + 1
		expected := upgradetypes.Plan{Name: exported.EmergencyHaltPlanName(haltHeight), Height: haltHeight, Info: "incident"}
		require.Len(t, upgradeK.ScheduleUpgradeCalls(), 1)
		assert.Equal(t, expected, upgradeK.ScheduleUpgradeCalls()[0].Plan)
		assert.Equal(t, &types.HaltChainResponse{HaltHeight: haltHeight, PlanName: expected.Name}, res)
		assert.Regexp(t, `^emergency-halt-\d+$`, expected.Name)

		event, ok := findEvent(ctx, types.EventTypeEmergencyHalt)
		require.True(t, ok)
		assert.Equal(t, sender, event[types.AttributeKeySigner])
		assert.Equal(t, "incident", event[types.AttributeKeyReason])
		assert.Equal(t, expected.Name, event[types.AttributeKeyPlanName])
		assert.NotContains(t, event, types.AttributeKeyReplacedPlanName)
	})

	t.Run("rejects a pending plan without override", func(t *testing.T) {
		ctx, _, server, upgradeK := setupPauseTest(t)
		pending := upgradetypes.Plan{Name: "v1.6", Height: ctx.BlockHeight() + 100, Info: "release"}
		upgradeK.GetUpgradePlanFunc = func(context.Context) (upgradetypes.Plan, error) { return pending, nil }

		_, err := server.HaltChain(ctx, &types.HaltChainRequest{Sender: rand.AccAddr().String(), Reason: "incident"})
		assert.ErrorContains(t, err, "upgrade plan v1.6 is already scheduled")
		assert.Empty(t, upgradeK.ScheduleUpgradeCalls())
	})

	t.Run("replaces a pending plan with override", func(t *testing.T) {
		ctx, _, server, upgradeK := setupPauseTest(t)
		pending := upgradetypes.Plan{Name: "v1.6", Height: ctx.BlockHeight() + 100, Info: "release"}
		upgradeK.GetUpgradePlanFunc = func(context.Context) (upgradetypes.Plan, error) { return pending, nil }

		_, err := server.HaltChain(ctx, &types.HaltChainRequest{Sender: rand.AccAddr().String(), Reason: "incident", OverridePendingUpgrade: true})
		require.NoError(t, err)
		require.Len(t, upgradeK.ScheduleUpgradeCalls(), 1)
		assert.Equal(t, exported.EmergencyHaltPlanName(ctx.BlockHeight()+1), upgradeK.ScheduleUpgradeCalls()[0].Plan.Name)

		event, ok := findEvent(ctx, types.EventTypeEmergencyHalt)
		require.True(t, ok)
		assert.Equal(t, "v1.6", event[types.AttributeKeyReplacedPlanName])
		assert.Equal(t, "release", event[types.AttributeKeyReplacedPlanInfo])
	})

	t.Run("fails if the upgrade keeper fails", func(t *testing.T) {
		ctx, _, server, upgradeK := setupPauseTest(t)
		upgradeK.ScheduleUpgradeFunc = func(context.Context, upgradetypes.Plan) error { return assert.AnError }

		_, err := server.HaltChain(ctx, &types.HaltChainRequest{Sender: rand.AccAddr().String(), Reason: "incident"})
		assert.ErrorIs(t, err, assert.AnError)
		_, ok := findEvent(ctx, types.EventTypeEmergencyHalt)
		assert.False(t, ok)
	})
}

func TestGenesis_PauseLevelRoundTrip(t *testing.T) {
	for _, level := range []exported.PauseLevel{exported.PAUSE_LEVEL_UNSPECIFIED, exported.PAUSE_LEVEL_ALL, exported.PAUSE_LEVEL_EXCEPT_VALIDATORS} {
		t.Run(level.String(), func(t *testing.T) {
			ctx, k, _, _ := setupPauseTest(t)

			genState := types.DefaultGenesisState()
			genState.PauseLevel = level
			require.NoError(t, genState.Validate())

			k.InitGenesis(ctx, genState)
			assert.Equal(t, level, k.GetPauseLevel(ctx))

			exported := k.ExportGenesis(ctx)
			assert.Equal(t, level, exported.PauseLevel)

			ctx2, k2, _, _ := setupPauseTest(t)
			k2.InitGenesis(ctx2, exported)
			assert.Equal(t, level, k2.GetPauseLevel(ctx2))
		})
	}
}

func TestQuery_IsPaused(t *testing.T) {
	ctx, k, server, _ := setupPauseTest(t)

	res, err := k.IsPaused(ctx, &types.IsPausedRequest{})
	require.NoError(t, err)
	assert.Equal(t, &types.IsPausedResponse{Paused: false, Level: exported.PAUSE_LEVEL_UNSPECIFIED}, res)

	_, err = server.Pause(ctx, &types.PauseRequest{Sender: rand.AccAddr().String(), Level: exported.PAUSE_LEVEL_EXCEPT_VALIDATORS, Reason: "incident"})
	require.NoError(t, err)

	res, err = k.IsPaused(ctx, &types.IsPausedRequest{})
	require.NoError(t, err)
	assert.Equal(t, &types.IsPausedResponse{Paused: true, Level: exported.PAUSE_LEVEL_EXCEPT_VALIDATORS}, res)
}

func TestCircuitBreaker_IsAllowed(t *testing.T) {
	ctx, k, server, _ := setupPauseTest(t)
	breaker := keeper.NewCircuitBreaker(k)
	bankSend := sdk.MsgTypeURL(&banktypes.MsgSend{})
	unpause := sdk.MsgTypeURL(&types.UnpauseRequest{})

	allowed, err := breaker.IsAllowed(ctx, bankSend)
	assert.NoError(t, err)
	assert.True(t, allowed)

	for _, level := range []exported.PauseLevel{exported.PAUSE_LEVEL_ALL, exported.PAUSE_LEVEL_EXCEPT_VALIDATORS} {
		_, err = server.Pause(ctx, &types.PauseRequest{Sender: rand.AccAddr().String(), Level: level, Reason: "incident"})
		require.NoError(t, err)

		allowed, err = breaker.IsAllowed(ctx, bankSend)
		assert.ErrorIs(t, err, types.ErrChainPaused)
		assert.False(t, allowed)

		for _, url := range exported.EmergencyPauseAllowlist[level] {
			allowed, err = breaker.IsAllowed(ctx, url)
			assert.NoError(t, err)
			assert.True(t, allowed, url)
		}

		allowed, err = breaker.IsAllowed(ctx, unpause)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}
}
