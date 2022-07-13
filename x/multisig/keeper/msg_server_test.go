package keeper_test

import (
	"errors"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper/mock"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	mock2 "github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"
)

func TestMsgServer(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	validators := slices.Expand(func(int) snapshot.Participant { return snapshot.NewParticipant(rand.ValAddr(), sdk.OneUint()) }, 10)

	var (
		msgServer   types.MsgServiceServer
		k           keeper.Keeper
		ctx         sdk.Context
		req         *types.SubmitPubKeyRequest
		snapshotter *mock.SnapshotterMock
		keyID       exported.KeyID
		expiresAt   int64
	)

	whenSenderIsProxy := When("the sender is a proxy", func() {
		snapshotter.GetOperatorFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return rand.Sample(validators, 1)[0].Address }
	})
	keySessionExists := When("a key session exists", func() {
		keyID = exported.KeyID(rand.HexStr(5))
		_, err := msgServer.StartKeygen(sdk.WrapSDKContext(ctx), types.NewStartKeygenRequest(rand.AccAddr(), keyID))
		expiresAt = ctx.BlockHeight() + types.DefaultParams().KeygenTimeout

		assert.NoError(t, err)
		assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, expiresAt), 1)
		assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()+types.DefaultParams().KeygenGracePeriod), 0)
	})
	requestIsMade := When("a request is made", func() {
		sk := funcs.Must(btcec.NewPrivateKey())
		req = types.NewSubmitPubKeyRequest(rand.AccAddr(), keyID, sk.PubKey().SerializeCompressed(), ecdsa.Sign(sk, []byte(keyID)).Serialize())
	})
	pubKeyFails := Then("submit pubkey fails", func(t *testing.T) {
		_, err := msgServer.SubmitPubKey(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
	})

	Given("a multisig msg server", func() {
		subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "multisig")
		k = keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)
		snapshotter = &mock.SnapshotterMock{
			CreateSnapshotFunc: func(sdk.Context, utils.Threshold) (snapshot.Snapshot, error) {
				return snapshot.NewSnapshot(ctx.BlockTime(), ctx.BlockHeight(), validators, sdk.NewUint(10)), nil
			},
		}
		msgServer = keeper.NewMsgServer(k, snapshotter, &mock2.StakerMock{})
	}).
		Given("a context", func() {
			ctx = rand2.Context(fake.NewMultiStore())
			k.InitGenesis(ctx, types.DefaultGenesisState())
		}).
		Branch(
			whenSenderIsProxy.
				When("the key ID does not exist", func() {
					// do not call StartKeygen
				}).
				When2(requestIsMade).
				Then2(pubKeyFails),

			When("the sender is not a proxy", func() {
				snapshotter.GetOperatorFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return nil }
			}).
				When2(keySessionExists).
				When2(requestIsMade).
				Then2(pubKeyFails),

			whenSenderIsProxy.
				When2(keySessionExists).
				When2(requestIsMade).
				Then("submit pubkey succeeds", func(t *testing.T) {
					_, err := msgServer.SubmitPubKey(sdk.WrapSDKContext(ctx), req)
					assert.NoError(t, err)
				}),

			whenSenderIsProxy.
				When("snapshot fails", func() {
					snapshotter.CreateSnapshotFunc = func(sdk.Context, utils.Threshold) (snapshot.Snapshot, error) {
						return snapshot.Snapshot{}, errors.New("some error")
					}
				}).
				Then("keygen fails", func(t *testing.T) {
					req := types.NewStartKeygenRequest(rand.AccAddr(), exported.KeyID(rand.HexStr(5)))
					_, err := msgServer.StartKeygen(sdk.WrapSDKContext(ctx), req)
					assert.Error(t, err)
				}),

			whenSenderIsProxy.
				When2(keySessionExists).
				Then("keygen with same KeyID fails", func(t *testing.T) {
					req := types.NewStartKeygenRequest(rand.AccAddr(), keyID)
					_, err := msgServer.StartKeygen(sdk.WrapSDKContext(ctx), req)
					assert.Error(t, err)
				}),

			whenSenderIsProxy.
				When("key exists", func() {
					k.SetKey(ctx, types.Key{
						ID:               keyID,
						Snapshot:         snapshot.NewSnapshot(ctx.BlockTime(), ctx.BlockHeight(), validators, sdk.NewUint(10)),
						SigningThreshold: types.DefaultParams().SigningThreshold,
					})
				}).
				Then("keygen with same KeyID fails", func(t *testing.T) {
					req := types.NewStartKeygenRequest(rand.AccAddr(), keyID)
					_, err := msgServer.StartKeygen(sdk.WrapSDKContext(ctx), req)
					assert.Error(t, err)
				}),

			keySessionExists.
				When("all participants submitted the public keys and the grace period does not go beyond the expires at", func() {
					for _, v := range validators {
						snapshotter.GetOperatorFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return v.Address }

						sk := funcs.Must(btcec.NewPrivateKey())
						req = types.NewSubmitPubKeyRequest(rand.AccAddr(), keyID, sk.PubKey().SerializeCompressed(), ecdsa.Sign(sk, []byte(keyID)).Serialize())

						_, err := msgServer.SubmitPubKey(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}
				}).
				Then("should update the keygen expiry", func(t *testing.T) {
					assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, expiresAt), 0)
					assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()+types.DefaultParams().KeygenGracePeriod+1), 1)
					assert.Equal(t, keyID, k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()+types.DefaultParams().KeygenGracePeriod+1)[0].GetKeyID())
				}),

			keySessionExists.
				When("all participants submitted the public keys and the grace period goes beyond the expires at", func() {
					ctx = ctx.WithBlockHeight(ctx.BlockHeight() + types.DefaultParams().KeygenTimeout - types.DefaultParams().KeygenGracePeriod)

					for _, v := range validators {
						snapshotter.GetOperatorFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return v.Address }

						sk := funcs.Must(btcec.NewPrivateKey())
						req = types.NewSubmitPubKeyRequest(rand.AccAddr(), keyID, sk.PubKey().SerializeCompressed(), ecdsa.Sign(sk, []byte(keyID)).Serialize())

						_, err := msgServer.SubmitPubKey(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}
				}).
				Then("should not update the keygen expiry if the grace period goes beyond expires at", func(t *testing.T) {
					assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, expiresAt), 1)
					assert.Equal(t, keyID, k.GetKeygenSessionsByExpiry(ctx, expiresAt)[0].GetKeyID())
				}),
		).Run(t)

}
