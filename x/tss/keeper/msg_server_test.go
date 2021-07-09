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
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
)

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
			AssignNextKeyFunc: func(sdk.Context, nexus.Chain, exported.KeyRole, string) error { return nil },
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
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (string, bool) { return "", false }
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (string, bool) { return "", true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.Bytes(sdk.AddrLen),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   rand.StrBetween(5, 20),
		})

		assert.NoError(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 1)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 1)
	}).Repeat(repeats))

	t.Run("next key is assigned correctly", testutils.Func(func(t *testing.T) {
		setup()
		keyID := rand.StrBetween(5, 20)
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (string, bool) { return rand.StrBetween(5, 20), true }
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (string, bool) { return keyID, true }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.Bytes(sdk.AddrLen),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   keyID,
		})

		assert.NoError(t, err)
		assert.Len(t, tssKeeper.AssignNextKeyCalls(), 0)
		assert.Len(t, tssKeeper.RotateKeyCalls(), 1)
	}).Repeat(repeats))

	t.Run("no next key is assigned", testutils.Func(func(t *testing.T) {
		setup()
		tssKeeper.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (string, bool) { return rand.StrBetween(5, 20), true }
		tssKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.Chain, exported.KeyRole) (string, bool) { return "", false }

		_, err := server.RotateKey(sdk.WrapSDKContext(ctx), &types.RotateKeyRequest{
			Sender:  rand.Bytes(sdk.AddrLen),
			Chain:   rand.StrBetween(5, 20),
			KeyRole: exported.KeyRole(rand.I64Between(1, 3)),
			KeyID:   rand.StrBetween(5, 20),
		})

		assert.Error(t, err)
	}).Repeat(repeats))
}
