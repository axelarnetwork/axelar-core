package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

func TestMsgServer_Ack(t *testing.T) {
	var (
		server      types.MsgServiceServer
		ctx         sdk.Context
		tssKeeper   *mock.TSSKeeperMock
		proxy       sdk.AccAddress
		validator   sdk.ValAddress
		eventHeight int64
	)
	setup := func() {
		proxy = rand.AccAddr()
		validator = rand.ValAddr()
		eventHeight = rand.I64Between(1, 5) * types.DefaultParams().AckPeriodInBlocks

		tssKeeper = &mock.TSSKeeperMock{
			LoggerFunc: func(_ sdk.Context) log.Logger { return ctx.Logger() },
			IsOperatorAvailableFunc: func(_ sdk.Context, v sdk.ValAddress) bool {
				return false
			},
			GetAckPeriodInBlocksFunc: func(_ sdk.Context) int64 {
				return types.DefaultParams().AckPeriodInBlocks
			},
			GetAckWindowInBlocksFunc: func(_ sdk.Context) int64 {
				return types.DefaultParams().AckWindowInBlocks
			},
			SetAvailableOperatorFunc: func(sdk.Context, sdk.ValAddress, int64) {},
		}
		snapshotter := &mock.SnapshotterMock{
			GetOperatorFunc: func(_ sdk.Context, p sdk.AccAddress) sdk.ValAddress {
				if p.Equals(proxy) {
					return validator
				}
				return sdk.ValAddress{}
			},
		}
		staker := &mock.StakingKeeperMock{}
		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{}

		server = NewMsgServerImpl(tssKeeper, snapshotter, staker, voter, nexusKeeper)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: eventHeight + rand.I64Between(1, types.DefaultParams().AckWindowInBlocks)}, false, log.TestingLogger())
	}

	repeats := 20

	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.Ack(sdk.WrapSDKContext(ctx), &types.AckRequest{
			Sender: proxy,
			Height: eventHeight,
		})

		assert.NoError(t, err)
		assert.Len(t, tssKeeper.IsOperatorAvailableCalls(), 1)
		assert.Len(t, tssKeeper.GetAckPeriodInBlocksCalls(), 1)
		assert.Len(t, tssKeeper.GetAckWindowInBlocksCalls(), 1)
		assert.Len(t, tssKeeper.SetAvailableOperatorCalls(), 1)
		assert.Len(t, tssKeeper.LoggerCalls(), 1)

	}).Repeat(repeats))

	t.Run("ACK already sent", testutils.Func(func(t *testing.T) {
		setup()

		tssKeeper.IsOperatorAvailableFunc = func(_ sdk.Context, v sdk.ValAddress) bool {
			return v.Equals(validator)
		}
		_, err := server.Ack(sdk.WrapSDKContext(ctx), &types.AckRequest{
			Sender: proxy,
			Height: eventHeight,
		})

		assert.Error(t, err)
		assert.Len(t, tssKeeper.IsOperatorAvailableCalls(), 1)
		assert.Len(t, tssKeeper.GetAckPeriodInBlocksCalls(), 0)
		assert.Len(t, tssKeeper.GetAckWindowInBlocksCalls(), 0)
		assert.Len(t, tssKeeper.SetAvailableOperatorCalls(), 0)
		assert.Len(t, tssKeeper.LoggerCalls(), 0)

	}).Repeat(repeats))

	t.Run("height mismatch", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.Ack(sdk.WrapSDKContext(ctx), &types.AckRequest{
			Sender: proxy,
			Height: eventHeight - types.DefaultParams().AckPeriodInBlocks,
		})

		assert.Error(t, err)
		assert.Len(t, tssKeeper.IsOperatorAvailableCalls(), 1)
		assert.Len(t, tssKeeper.GetAckPeriodInBlocksCalls(), 1)
		assert.Len(t, tssKeeper.GetAckWindowInBlocksCalls(), 0)
		assert.Len(t, tssKeeper.SetAvailableOperatorCalls(), 0)
		assert.Len(t, tssKeeper.LoggerCalls(), 0)

	}).Repeat(repeats))

	t.Run("late ACK", testutils.Func(func(t *testing.T) {
		setup()

		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + types.DefaultParams().AckWindowInBlocks + rand.I64Between(1, 10))

		_, err := server.Ack(sdk.WrapSDKContext(ctx), &types.AckRequest{
			Sender: proxy,
			Height: eventHeight,
		})

		assert.NoError(t, err)
		assert.Len(t, tssKeeper.IsOperatorAvailableCalls(), 1)
		assert.Len(t, tssKeeper.GetAckPeriodInBlocksCalls(), 1)
		assert.Len(t, tssKeeper.GetAckWindowInBlocksCalls(), 2)
		assert.Len(t, tssKeeper.SetAvailableOperatorCalls(), 0)
		assert.Len(t, tssKeeper.LoggerCalls(), 1)

	}).Repeat(repeats))
}

func TestMsgServer_RotateKey(t *testing.T) {
	var (
		server    types.MsgServiceServer
		ctx       sdk.Context
		tssKeeper *mock.TSSKeeperMock
	)
	setup := func() {
		tssKeeper = &mock.TSSKeeperMock{
			RotateKeyFunc:     func(sdk.Context, nexus.Chain, exported.KeyRole) error { return nil },
			LoggerFunc:        func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			AssignNextKeyFunc: func(sdk.Context, nexus.Chain, exported.KeyRole, exported.KeyID) error { return nil },
			AssertMatchesRequirementsFunc: func(ctx sdk.Context, snapshotter snapshot.Snapshotter, chain nexus.Chain, keyID exported.KeyID, keyRole exported.KeyRole) error {
				return nil
			},
		}
		snapshotter := &mock.SnapshotterMock{}
		staker := &mock.StakingKeeperMock{}
		voter := &mock.VoterMock{}
		nexusKeeper := &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 10),
					SupportsForeignAssets: rand.Bools(0.5).Next(),
				}, true
			},
		}
		server = NewMsgServerImpl(tssKeeper, snapshotter, staker, voter, nexusKeeper)
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	}
	repeats := 20
	t.Run("first key rotation", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", false }
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   tssTestUtils.RandKeyID(),
		})

		assert.NoError(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 1)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 1)
	}).Repeat(repeats))

	t.Run("next key is assigned", testutils.Func(func(t *testing.T) {
		setup()
		keyID := tssTestUtils.RandKeyID()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return keyID, true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   keyID,
		})

		assert.Error(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 0)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 0)
	}).Repeat(repeats))

	t.Run("no next key is assigned", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (exported.KeyID, bool) { return "", false }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.AccAddr(),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   tssTestUtils.RandKeyID(),
		})

		assert.Error(t, err)
	}).Repeat(repeats))
}
