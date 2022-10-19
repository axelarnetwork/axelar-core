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
	exportedmock "github.com/axelarnetwork/axelar-core/x/multisig/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper/mock"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	mock2 "github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
	"github.com/axelarnetwork/utils/test/rand"
)

func TestMsgServer(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	validators := slices.Expand(func(int) snapshot.Participant { return snapshot.NewParticipant(rand2.ValAddr(), sdk.OneUint()) }, 10)

	var (
		msgServer   types.MsgServiceServer
		k           keeper.Keeper
		ctx         sdk.Context
		req         *types.SubmitPubKeyRequest
		snapshotter *mock.SnapshotterMock
		nexusK      *mock2.NexusMock
		keyID       exported.KeyID
		expiresAt   int64
	)

	givenMsgServer := Given("a multisig msg server", func() {
		subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "multisig")
		k = keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace)

		ctx = rand2.Context(fake.NewMultiStore())
		k.InitGenesis(ctx, types.DefaultGenesisState())
		snapshotter = &mock.SnapshotterMock{
			CreateSnapshotFunc: func(sdk.Context, utils.Threshold) (snapshot.Snapshot, error) {
				return snapshot.NewSnapshot(ctx.BlockTime(), ctx.BlockHeight(), validators, sdk.NewUint(10)), nil
			},
		}
		nexusK = &mock2.NexusMock{}

		msgServer = keeper.NewMsgServer(k, snapshotter, &mock2.StakerMock{}, nexusK)
	})

	whenSenderIsProxy := When("the sender is a proxy", func() {
		snapshotter.GetOperatorFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return rand.Sample(validators, 1)[0].Address }
	})
	keySessionExists := When("a key session exists", func() {
		keyID = exported.KeyID(rand.HexStr(5))
		_, err := msgServer.StartKeygen(sdk.WrapSDKContext(ctx), types.NewStartKeygenRequest(rand2.AccAddr(), keyID))
		expiresAt = ctx.BlockHeight() + types.DefaultParams().KeygenTimeout

		assert.NoError(t, err)
		assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, expiresAt), 1)
		assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()+types.DefaultParams().KeygenGracePeriod), 0)
	})
	requestIsMade := When("a request is made", func() {
		sk := funcs.Must(btcec.NewPrivateKey())
		req = types.NewSubmitPubKeyRequest(rand2.AccAddr(), keyID, sk.PubKey().SerializeCompressed(), ecdsa.Sign(sk, []byte(keyID)).Serialize())
	})
	pubKeyFails := Then("submit pubkey fails", func(t *testing.T) {
		_, err := msgServer.SubmitPubKey(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
	})

	t.Run("keygen", func(t *testing.T) {
		givenMsgServer.
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
						req := types.NewStartKeygenRequest(rand2.AccAddr(), exported.KeyID(rand.HexStr(5)))
						_, err := msgServer.StartKeygen(sdk.WrapSDKContext(ctx), req)
						assert.Error(t, err)
					}),

				whenSenderIsProxy.
					When2(keySessionExists).
					Then("keygen with same KeyID fails", func(t *testing.T) {
						req := types.NewStartKeygenRequest(rand2.AccAddr(), keyID)
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
						req := types.NewStartKeygenRequest(rand2.AccAddr(), keyID)
						_, err := msgServer.StartKeygen(sdk.WrapSDKContext(ctx), req)
						assert.Error(t, err)
					}),

				keySessionExists.
					When("all participants submitted the public keys and the grace period does not go beyond the expires at", func() {
						for _, v := range validators {
							snapshotter.GetOperatorFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return v.Address }

							sk := funcs.Must(btcec.NewPrivateKey())
							req = types.NewSubmitPubKeyRequest(rand2.AccAddr(), keyID, sk.PubKey().SerializeCompressed(), ecdsa.Sign(sk, []byte(keyID)).Serialize())

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
							req = types.NewSubmitPubKeyRequest(rand2.AccAddr(), keyID, sk.PubKey().SerializeCompressed(), ecdsa.Sign(sk, []byte(keyID)).Serialize())

							_, err := msgServer.SubmitPubKey(sdk.WrapSDKContext(ctx), req)
							assert.NoError(t, err)
						}
					}).
					Then("should not update the keygen expiry if the grace period goes beyond expires at", func(t *testing.T) {
						assert.Len(t, k.GetKeygenSessionsByExpiry(ctx, expiresAt), 1)
						assert.Equal(t, keyID, k.GetKeygenSessionsByExpiry(ctx, expiresAt)[0].GetKeyID())
					}),
			).Run(t)
	})

	t.Run("signing", func(t *testing.T) {
		var (
			payloadHash exported.Hash
			sigID       uint64
			key         types.Key
		)

		participantCount := 3
		validators := slices.Expand(func(int) sdk.ValAddress { return rand2.ValAddr() }, participantCount)
		proxies := slices.Expand(func(int) sdk.AccAddress { return rand2.AccAddr() }, participantCount)
		participants := slices.Map(validators, func(v sdk.ValAddress) snapshot.Participant { return snapshot.NewParticipant(v, sdk.OneUint()) })
		keyID := exported.KeyID(rand.HexStr(5))
		privateKeys := slices.Expand(func(int) *btcec.PrivateKey { return funcs.Must(btcec.NewPrivateKey()) }, participantCount)
		publicKeys := slices.Map(privateKeys, func(sk *btcec.PrivateKey) exported.PublicKey { return sk.PubKey().SerializeCompressed() })
		module := rand.AlphaStrBetween(3, 3)

		givenMsgServer.
			When("proxies are all set up", func() {
				snapshotter.GetOperatorFunc = func(_ sdk.Context, p sdk.AccAddress) sdk.ValAddress {
					for i, proxy := range proxies {
						if proxy.Equals(p) {
							return validators[i]
						}
					}

					return nil
				}
			}).
			When("sig handler is set", func() {
				sigRouter := types.NewSigRouter().AddHandler(module, &exportedmock.SigHandlerMock{})
				k.SetSigRouter(sigRouter)
			}).
			When("key is generated", func() {
				pubKeyIndex := 0
				key = types.Key{
					ID:       keyID,
					Snapshot: snapshot.NewSnapshot(ctx.BlockTime(), ctx.BlockHeight(), participants, sdk.NewUint(uint64(participantCount))),
					PubKeys: slices.ToMap(publicKeys, func(pk exported.PublicKey) string {
						result := validators[pubKeyIndex]
						pubKeyIndex++

						return result.String()
					}),
					SigningThreshold: utils.NewThreshold(2, 3),
				}

				k.SetKey(ctx, key)
			}).
			Branch(
				Then("should panic if sig handler is not registered for the given module", func(t *testing.T) {
					assert.Panics(t, func() {
						k.Sign(ctx, exported.KeyID(rand.HexStr(5)), rand.Bytes(exported.HashLength), rand.AlphaStrBetween(3, 3))
					})
				}),

				Then("should fail if the key does not exist", func(t *testing.T) {
					err := k.Sign(ctx, exported.KeyID(rand.HexStr(5)), rand.Bytes(exported.HashLength), module)

					assert.Error(t, err)
				}),

				Then("should fail if the key is inactive", func(t *testing.T) {
					err := k.Sign(ctx, keyID, rand.Bytes(exported.HashLength), module)

					assert.ErrorContains(t, err, "not activated")
				}),

				Then("should fail if the key is assigned", func(t *testing.T) {
					key.State = exported.Assigned
					k.SetKey(ctx, key)

					err := k.Sign(ctx, keyID, rand.Bytes(exported.HashLength), module)

					assert.ErrorContains(t, err, "not activated")
				}),

				When("key is active", func() {
					key.State = exported.Active
					k.SetKey(ctx, key)
				}).Branch(
					Then("should fail if payload hash is invalid", func(t *testing.T) {
						err := k.Sign(ctx, keyID, rand.Bytes(100), module)
						assert.Error(t, err)

						var zeroHash [exported.HashLength]byte
						err = k.Sign(ctx, keyID, zeroHash[:], module)
						assert.Error(t, err)
					}),

					Then("should start signing if the key exists", func(t *testing.T) {
						err := k.Sign(ctx, keyID, rand.Bytes(exported.HashLength), module)

						assert.NoError(t, err)
						assert.Len(t, k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()+types.DefaultParams().SigningTimeout), 1)
					}),

					When("signing session exists", func() {
						payloadHash = rand.Bytes(exported.HashLength)
						funcs.MustNoErr(k.Sign(ctx, keyID, payloadHash, module))

						events := ctx.EventManager().Events().ToABCIEvents()
						sigID = funcs.Must(sdk.ParseTypedEvent(events[len(events)-1])).(*types.SigningStarted).SigID
					}).
						Branch(
							Then("should fail if the signing session does not exist", func(t *testing.T) {
								pIndex := rand.I64Between(1, participantCount)
								signature := ecdsa.Sign(privateKeys[pIndex], payloadHash).Serialize()
								_, err := msgServer.SubmitSignature(sdk.WrapSDKContext(ctx), types.NewSubmitSignatureRequest(proxies[pIndex], uint64(rand.PosI64()), signature))

								assert.Error(t, err)
							}),

							Then("should fail if proxy is not registered", func(t *testing.T) {
								signature := ecdsa.Sign(privateKeys[rand.I64Between(1, participantCount)], payloadHash).Serialize()
								_, err := msgServer.SubmitSignature(sdk.WrapSDKContext(ctx), types.NewSubmitSignatureRequest(rand2.AccAddr(), sigID, signature))

								assert.Error(t, err)
							}),

							Then("should succeed", func(t *testing.T) {
								for i, proxy := range proxies {
									signature := ecdsa.Sign(privateKeys[i], payloadHash).Serialize()
									_, err := msgServer.SubmitSignature(sdk.WrapSDKContext(ctx), types.NewSubmitSignatureRequest(proxy, sigID, signature))

									assert.NoError(t, err)
								}

								assert.Len(t, k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()+types.DefaultParams().SigningTimeout), 0)
								actual := k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()+types.DefaultParams().SigningGracePeriod+1)
								assert.Len(t, actual, 1)

								sig, err := actual[0].Result()
								assert.NoError(t, err)
								assert.NoError(t, sig.ValidateBasic())

								participantsWeight := sdk.ZeroUint()
								for p := range sig.GetSigs() {
									participantsWeight = participantsWeight.Add(key.GetSnapshot().GetParticipantWeight(funcs.Must(sdk.ValAddressFromBech32(p))))
								}

								assert.True(t, participantsWeight.GTE(key.GetMinPassingWeight()))
							}),
						),
				),
			).
			Run(t)
	})

	t.Run("RotateKey", func(t *testing.T) {
		var (
			keyID exported.KeyID
			chain nexus.ChainName
		)

		givenMsgServer.
			When("key is generated", func() {
				key := testutils.Key()
				keyID = key.GetID()

				k.SetKey(ctx, key)
			}).
			Branch(
				When("chain is unknown", func() {
					nexusK.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, false }
				}).
					Then("should fail", func(t *testing.T) {
						_, err := msgServer.RotateKey(sdk.WrapSDKContext(ctx), types.NewRotateKeyRequest(rand2.AccAddr(), nexus.ChainName(rand.AlphaStrBetween(1, 5)), keyID))
						assert.Error(t, err)
					}),

				When("chain is known", func() {
					chain = nexus.ChainName(rand.AlphaStrBetween(1, 5))
					nexusK.GetChainFunc = func(ctx sdk.Context, cn nexus.ChainName) (nexus.Chain, bool) {
						return nexus.Chain{}, cn == chain
					}
				}).
					Then("should succeed but only once", func(t *testing.T) {
						_, err := msgServer.RotateKey(sdk.WrapSDKContext(ctx), types.NewRotateKeyRequest(rand2.AccAddr(), chain, keyID))
						assert.NoError(t, err)

						_, err = msgServer.RotateKey(sdk.WrapSDKContext(ctx), types.NewRotateKeyRequest(rand2.AccAddr(), chain, keyID))
						assert.Error(t, err)
					}),
			).
			Run(t)
	})
}
