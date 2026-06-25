package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmkeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	rewardkeeper "github.com/axelarnetwork/axelar-core/x/reward/keeper"
	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	rewardmock "github.com/axelarnetwork/axelar-core/x/reward/types/mock"
)

func TestMsgServerActivateDeactivateWasm(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "nexus")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	snap := mock.SnapshotterMock{}
	slashing := mock.SlashingKeeperMock{}
	staking := mock.StakingKeeperMock{}
	ax := mock.AxelarnetKeeperMock{}

	msgServer := keeper.NewMsgServerImpl(k, &snap, &slashing, &staking, &ax, &mock.RewardKeeperMock{})

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))

	assert.True(t, k.IsWasmConnectionActivated(ctx))

	_, err := msgServer.DeactivateChain(sdk.WrapSDKContext(ctx), &types.DeactivateChainRequest{Chains: []nexus.ChainName{":all:"}})
	assert.NoError(t, err)
	assert.False(t, k.IsWasmConnectionActivated(ctx))

	_, err = msgServer.ActivateChain(sdk.WrapSDKContext(ctx), &types.ActivateChainRequest{Chains: []nexus.ChainName{":wasm:"}})
	assert.NoError(t, err)
	assert.True(t, k.IsWasmConnectionActivated(ctx))

	_, err = msgServer.DeactivateChain(sdk.WrapSDKContext(ctx), &types.DeactivateChainRequest{Chains: []nexus.ChainName{"not_wasm"}})
	assert.NoError(t, err)
	assert.True(t, k.IsWasmConnectionActivated(ctx))
}

func TestUpdateParams(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "nexus")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	msgServer := keeper.NewMsgServerImpl(k, &mock.SnapshotterMock{}, &mock.SlashingKeeperMock{}, &mock.StakingKeeperMock{}, &mock.AxelarnetKeeperMock{}, &mock.RewardKeeperMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))

	p := types.DefaultParams()
	p.ChainMaintainerCheckWindow = p.ChainMaintainerCheckWindow + 1
	_, err := msgServer.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: "", Params: p})
	assert.NoError(t, err)
	assert.Equal(t, p, k.GetParams(ctx))
}

func TestRetryFailedMessage(t *testing.T) {
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "nexus")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
	)

	validators := types.NewAddressValidators()
	validators.AddAddressValidator(evmTypes.ModuleName, evmkeeper.NewAddressValidator())
	validators.Seal()
	k.SetAddressValidators(validators)

	msgServer := keeper.NewMsgServerImpl(k, &mock.SnapshotterMock{}, &mock.SlashingKeeperMock{}, &mock.StakingKeeperMock{}, &mock.AxelarnetKeeperMock{}, &mock.RewardKeeperMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))

	k.SetMessageRouter(types.NewMessageRouter().AddRoute(evm.Ethereum.Module, func(_ sdk.Context, _ nexus.RoutingContext, _ nexus.GeneralMessage) error {
		return nil
	}))

	t.Run("message not found", func(t *testing.T) {
		_, err := msgServer.RetryFailedMessage(ctx, &types.RetryFailedMessageRequest{
			Sender: rand.AccAddr().String(),
			ID:     "non-existent-id",
		})
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("message not in failed status", func(t *testing.T) {
		msg := nexus.GeneralMessage{
			ID: rand.NormalizedStr(10),
			Sender: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			Recipient: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			PayloadHash:   evmtestutils.RandomHash().Bytes(),
			Status:        nexus.Approved,
			SourceTxID:    evmtestutils.RandomHash().Bytes(),
			SourceTxIndex: 0,
		}

		err := k.SetNewMessage(ctx, msg)
		assert.NoError(t, err)

		_, err = msgServer.RetryFailedMessage(ctx, &types.RetryFailedMessageRequest{
			Sender: rand.AccAddr().String(),
			ID:     msg.ID,
		})
		assert.ErrorContains(t, err, "not in failed status")
	})

	t.Run("successfully retry failed message", func(t *testing.T) {
		k.SetChain(ctx, evm.Ethereum)
		k.ActivateChain(ctx, evm.Ethereum)

		msg := nexus.GeneralMessage{
			ID: rand.NormalizedStr(10),
			Sender: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			Recipient: nexus.CrossChainAddress{
				Chain:   evm.Ethereum,
				Address: evmtestutils.RandomAddress().Hex(),
			},
			PayloadHash:   evmtestutils.RandomHash().Bytes(),
			Status:        nexus.Approved,
			SourceTxID:    evmtestutils.RandomHash().Bytes(),
			SourceTxIndex: 0,
		}

		err := k.SetNewMessage(ctx, msg)
		assert.NoError(t, err)

		err = k.RouteMessage(ctx, msg.ID)
		assert.NoError(t, err)

		err = k.SetMessageFailed(ctx, msg.ID)
		assert.NoError(t, err)

		storedMsg, ok := k.GetMessage(ctx, msg.ID)
		assert.True(t, ok)
		assert.Equal(t, nexus.Failed, storedMsg.Status)

		_, err = msgServer.RetryFailedMessage(ctx, &types.RetryFailedMessageRequest{
			Sender: rand.AccAddr().String(),
			ID:     msg.ID,
		})
		assert.NoError(t, err)

		dequeuedMsg, ok := k.DequeueRouteMessage(ctx)
		assert.True(t, ok)
		assert.Equal(t, msg.ID, dequeuedMsg.ID)
	})
}

func TestDeregisterChainMaintainerClearsRewards(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encCfg.Amino)
	types.RegisterInterfaces(encCfg.InterfaceRegistry)

	chain := evm.Ethereum
	poolName := chain.Name.String()
	accrued := sdk.NewCoin("uaxl", math.NewInt(1_000_000))

	type fixture struct {
		ctx          sdk.Context
		rewardKeeper rewardkeeper.Keeper
		banker       *rewardmock.BankerMock
		validator    sdk.ValAddress
		register     func()
		deregister   func()
	}

	setup := func(t *testing.T) fixture {
		ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))
		validator := rand.ValAddr()
		sender := rand.AccAddr()

		// real reward keeper so we exercise the actual reward-pool storage
		banker := &rewardmock.BankerMock{
			MintCoinsFunc:                   func(context.Context, string, sdk.Coins) error { return nil },
			SendCoinsFromModuleToModuleFunc: func(context.Context, string, string, sdk.Coins) error { return nil },
		}
		distributor := &rewardmock.DistributorMock{
			AllocateTokensToValidatorFunc: func(context.Context, stakingtypes.ValidatorI, sdk.DecCoins) error { return nil },
		}
		staker := &rewardmock.StakerMock{
			ValidatorFunc: func(context.Context, sdk.ValAddress) (stakingtypes.ValidatorI, error) {
				return stakingtypes.Validator{}, nil
			},
		}
		rewardSubspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("rewardParams"), store.NewKVStoreKey("rewardTParams"), rewardtypes.ModuleName)
		rewardKeeper := rewardkeeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey(rewardtypes.StoreKey), rewardSubspace, banker, distributor, staker)

		// real nexus keeper so register/deregister mutate actual maintainer state
		nexusSubspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("nexusParams"), store.NewKVStoreKey("nexusTParams"), types.ModuleName)
		nexusKeeper := keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey(types.StoreKey), nexusSubspace)
		nexusKeeper.SetChain(ctx, chain)

		snap := &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return validator },
		}
		staking := &mock.StakingKeeperMock{
			ValidatorFunc: func(context.Context, sdk.ValAddress) (stakingtypes.ValidatorI, error) {
				return stakingtypes.Validator{Status: stakingtypes.Bonded}, nil
			},
		}
		ax := &mock.AxelarnetKeeperMock{
			IsCosmosChainFunc: func(sdk.Context, nexus.ChainName) bool { return false },
		}
		msgServer := keeper.NewMsgServerImpl(nexusKeeper, snap, &mock.SlashingKeeperMock{}, staking, ax, rewardKeeper)

		register := func() {
			t.Helper()
			_, err := msgServer.RegisterChainMaintainer(sdk.WrapSDKContext(ctx), types.NewRegisterChainMaintainerRequest(sender, chain.Name.String()))
			assert.NoError(t, err)
			assert.True(t, nexusKeeper.IsChainMaintainer(ctx, chain, validator))
		}
		deregister := func() {
			t.Helper()
			_, err := msgServer.DeregisterChainMaintainer(sdk.WrapSDKContext(ctx), types.NewDeregisterChainMaintainerRequest(sender, chain.Name.String()))
			assert.NoError(t, err)
			assert.False(t, nexusKeeper.IsChainMaintainer(ctx, chain, validator))
		}

		return fixture{ctx: ctx, rewardKeeper: rewardKeeper, banker: banker, validator: validator, register: register, deregister: deregister}
	}

	// control: an accrued reward is releasable while still registered, proving the
	// seeding/release path works and the exploit assertion below isn't vacuous.
	t.Run("accrued reward is releasable while still registered", func(t *testing.T) {
		f := setup(t)
		f.register()
		f.rewardKeeper.GetPool(f.ctx, poolName).AddReward(f.validator, accrued)

		assert.NoError(t, f.rewardKeeper.GetPool(f.ctx, poolName).ReleaseRewards(f.validator))
		assert.Len(t, f.banker.MintCoinsCalls(), 1)
	})

	// the fix: deregistering clears the accrued balance, so it cannot be preserved
	// across a re-registration and released later (the reported penalty dodge).
	t.Run("deregister clears the accrued reward across re-registration", func(t *testing.T) {
		f := setup(t)
		f.register()
		f.rewardKeeper.GetPool(f.ctx, poolName).AddReward(f.validator, accrued)

		f.deregister()
		f.register()

		assert.NoError(t, f.rewardKeeper.GetPool(f.ctx, poolName).ReleaseRewards(f.validator))
		assert.Empty(t, f.banker.MintCoinsCalls(), "rewards must be cleared on deregister, not preserved across re-registration")
	})
}
