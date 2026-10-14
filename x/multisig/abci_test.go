package multisig_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	exportedmock "github.com/axelarnetwork/axelar-core/x/multisig/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	typestestutils "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	rewardmock "github.com/axelarnetwork/axelar-core/x/reward/exported/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestEndBlocker(t *testing.T) {
	var (
		ctx           sdk.Context
		k             *mock.KeeperMock
		rewarder      *mock.RewarderMock
		keygenSession types.KeygenSession
		failingKeygen types.KeygenSession
		healthyKeygen types.KeygenSession
	)

	givenKeepersAndCtx := Given("keepers", func() {
		ctx = rand.Context(fake.NewMultiStore(), t)
		k = &mock.KeeperMock{
			LoggerFunc:                     func(sdk.Context) log.Logger { return log.NewTestLogger(t) },
			GetKeygenSessionsByExpiryFunc:  func(sdk.Context, int64) []types.KeygenSession { return nil },
			GetSigningSessionsByExpiryFunc: func(sdk.Context, int64) []types.SigningSession { return nil },
		}
		rewarder = &mock.RewarderMock{}
	})

	t.Run("handleKeygens", func(t *testing.T) {
		givenKeepersAndCtx.
			When("a pending keygen session expiry equal to the block height", func() {
				k.GetKeygenSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.KeygenSession {
					if expiry != ctx.BlockHeight()+1 {
						return nil
					}

					return []types.KeygenSession{{
						Key: types.Key{
							ID:       testutils.KeyID(),
							Snapshot: snapshottestutils.Snapshot(uint64(rand.I64Between(10, 11)), utils.NewThreshold(1, 2)),
						},
						State: exported.Pending,
					}}
				}
			}).
			Then("should delete and penalize missing participants", func(t *testing.T) {
				pool := rewardmock.RewardPoolMock{
					ClearRewardsFunc: func(sdk.ValAddress) {},
				}

				k.DeleteKeygenSessionFunc = func(sdk.Context, exported.KeyID) {}
				rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }

				_, err := multisig.EndBlocker(ctx, k, rewarder)

				assert.NoError(t, err)
				assert.Len(t, k.DeleteKeygenSessionCalls(), 1)
				assert.Len(t, pool.ClearRewardsCalls(), 10)
			}).
			Run(t, 20)

		givenKeepersAndCtx.
			When("a completed keygen session expiry equal to the block height", func() {
				keygenSession = types.KeygenSession{
					Key:   typestestutils.Key(),
					State: exported.Completed,
				}
				k.GetKeygenSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.KeygenSession {
					if expiry != ctx.BlockHeight()+1 {
						return nil
					}

					return []types.KeygenSession{keygenSession}
				}
			}).
			Then("should delete and set key", func(t *testing.T) {
				pool := rewardmock.RewardPoolMock{
					ReleaseRewardsFunc: func(sdk.ValAddress) error { return nil },
				}
				rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
				k.DeleteKeygenSessionFunc = func(sdk.Context, exported.KeyID) {}
				k.SetKeyFunc = func(sdk.Context, types.Key) {}

				_, err := multisig.EndBlocker(ctx, k, rewarder)

				assert.NoError(t, err)
				assert.Len(t, k.DeleteKeygenSessionCalls(), 1)
				assert.Len(t, k.SetKeyCalls(), 1)
				assert.Len(t, pool.ReleaseRewardsCalls(), len(keygenSession.Key.Snapshot.GetParticipantAddresses()))
			}).
			Run(t, 20)

		givenKeepersAndCtx.
			When("a completed keygen session with missing participants expiry equal to the block height", func() {
				key := typestestutils.KeyWithMissingParticipants()

				keygenSession = types.KeygenSession{
					Key:   key,
					State: exported.Completed,
				}
				k.GetKeygenSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.KeygenSession {
					if expiry != ctx.BlockHeight()+1 {
						return nil
					}

					return []types.KeygenSession{keygenSession}
				}
			}).
			Then("should delete and set key", func(t *testing.T) {
				pool := rewardmock.RewardPoolMock{
					ClearRewardsFunc:   func(sdk.ValAddress) {},
					ReleaseRewardsFunc: func(sdk.ValAddress) error { return nil },
				}
				rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
				k.DeleteKeygenSessionFunc = func(sdk.Context, exported.KeyID) {}
				k.SetKeyFunc = func(sdk.Context, types.Key) {}

				_, err := multisig.EndBlocker(ctx, k, rewarder)

				assert.NoError(t, err)
				assert.Len(t, k.DeleteKeygenSessionCalls(), 1)
				assert.Len(t, k.SetKeyCalls(), 1)
				missingCount := len(keygenSession.GetMissingParticipants())
				assert.Len(t, pool.ReleaseRewardsCalls(), len(keygenSession.Key.Snapshot.GetParticipantAddresses())-missingCount)
				assert.Len(t, pool.ClearRewardsCalls(), missingCount)
			}).
			Run(t, 20)

		givenKeepersAndCtx.
			When("a completed keygen session whose reward release fails", func() {
				keygenSession = types.KeygenSession{
					Key:   typestestutils.Key(),
					State: exported.Completed,
				}
				k.GetKeygenSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.KeygenSession {
					if expiry != ctx.BlockHeight()+1 {
						return nil
					}

					return []types.KeygenSession{keygenSession}
				}
			}).
			Then("recover, keep session cleanup and not set the key", func(t *testing.T) {
				pool := rewardmock.RewardPoolMock{
					ClearRewardsFunc:   func(sdk.ValAddress) {},
					ReleaseRewardsFunc: func(sdk.ValAddress) error { return fmt.Errorf("reward release failed") },
				}
				rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
				storeKey := store.NewKVStoreKey("cache")
				deletedKey := []byte("deleted")
				keySetKey := []byte("key-set")
				k.DeleteKeygenSessionFunc = func(ctx sdk.Context, _ exported.KeyID) {
					ctx.MultiStore().GetKVStore(storeKey).Set(deletedKey, []byte{})
				}
				k.SetKeyFunc = func(ctx sdk.Context, _ types.Key) {
					ctx.MultiStore().GetKVStore(storeKey).Set(keySetKey, []byte{})
				}

				assert.NotPanics(t, func() {
					_, err := multisig.EndBlocker(ctx, k, rewarder)
					assert.NoError(t, err)
				})
				assert.True(t, ctx.MultiStore().GetKVStore(storeKey).Has(deletedKey))
				assert.False(t, ctx.MultiStore().GetKVStore(storeKey).Has(keySetKey))
				assert.Len(t, k.DeleteKeygenSessionCalls(), 1)
			}).
			Run(t, 20)

		givenKeepersAndCtx.
			When("two completed keygen sessions where the first one fails", func() {
				failingKeygen = types.KeygenSession{
					Key:   typestestutils.Key(),
					State: exported.Completed,
				}
				healthyKeygen = types.KeygenSession{
					Key:   typestestutils.Key(),
					State: exported.Completed,
				}
				k.GetKeygenSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.KeygenSession {
					if expiry != ctx.BlockHeight()+1 {
						return nil
					}

					return []types.KeygenSession{failingKeygen, healthyKeygen}
				}
			}).
			Then("still process the second one", func(t *testing.T) {
				failingParticipants := make(map[string]bool)
				for _, p := range failingKeygen.Key.GetParticipants() {
					failingParticipants[p.String()] = true
				}

				pool := rewardmock.RewardPoolMock{
					ClearRewardsFunc: func(sdk.ValAddress) {},
					ReleaseRewardsFunc: func(p sdk.ValAddress) error {
						if failingParticipants[p.String()] {
							return fmt.Errorf("reward release failed")
						}
						return nil
					},
				}
				rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
				storeKey := store.NewKVStoreKey("cache")
				k.DeleteKeygenSessionFunc = func(sdk.Context, exported.KeyID) {}
				k.SetKeyFunc = func(ctx sdk.Context, key types.Key) {
					ctx.MultiStore().GetKVStore(storeKey).Set([]byte(key.ID), []byte{})
				}

				assert.NotPanics(t, func() {
					_, err := multisig.EndBlocker(ctx, k, rewarder)
					assert.NoError(t, err)
				})
				assert.True(t, ctx.MultiStore().GetKVStore(storeKey).Has([]byte(healthyKeygen.Key.ID)))
				assert.False(t, ctx.MultiStore().GetKVStore(storeKey).Has([]byte(failingKeygen.Key.ID)))
				assert.Len(t, k.DeleteKeygenSessionCalls(), 2)
			}).
			Run(t, 20)
	})

	t.Run("handleSignings", func(t *testing.T) {
		var (
			module         string
			sigHandler     *exportedmock.SigHandlerMock
			signingSession types.SigningSession
			missingCount   uint64
		)

		givenKeepersAndCtx.
			When("module sig handler is set", func() {
				module = rand.NormalizedStr(5)
				sigHandler = &exportedmock.SigHandlerMock{}

				k.GetSigRouterFunc = func() types.SigRouter {
					sigRouter := types.NewSigRouter()
					sigRouter.AddHandler(module, sigHandler)
					sigRouter.Seal()

					return sigRouter
				}
			}).
			Branch(
				When("a pending signing session expiry equal to the block height", func() {
					k.GetSigningSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.SigningSession {
						if expiry != ctx.BlockHeight()+1 {
							return nil
						}

						return []types.SigningSession{{
							ID:     uint64(rand.PosI64()),
							Module: module,
							Key: types.Key{
								ID: testutils.KeyID(),
								PubKeys: slices.ToMap(
									slices.Expand(func(int) exported.PublicKey { return funcs.Must(btcec.NewPrivateKey()).PubKey().SerializeCompressed() }, 10),
									func(exported.PublicKey) string { return rand.ValAddr().String() },
								),
							},
							State: exported.Pending,
						}}
					}
				}).
					Then("should delete and penalize missing participants", func(t *testing.T) {
						pool := rewardmock.RewardPoolMock{
							ClearRewardsFunc: func(sdk.ValAddress) {},
						}

						k.DeleteSigningSessionFunc = func(sdk.Context, uint64) {}
						rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
						sigHandler.HandleFailedFunc = func(sdk.Context, codec.ProtoMarshaler) error { return nil }

						_, err := multisig.EndBlocker(ctx, k, rewarder)

						assert.NoError(t, err)
						assert.Len(t, k.DeleteSigningSessionCalls(), 1)
						assert.Len(t, pool.ClearRewardsCalls(), 10)
						assert.Len(t, sigHandler.HandleFailedCalls(), 1)
					}),

				When("a completed signing session expiry equal to the block height", func() {
					signingSession = newSigningSession(module)
					k.GetSigningSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.SigningSession {
						if expiry != ctx.BlockHeight()+1 {
							return nil
						}

						return []types.SigningSession{signingSession}
					}
				}).
					Then("should delete and set sig", func(t *testing.T) {
						pool := rewardmock.RewardPoolMock{
							ReleaseRewardsFunc: func(sdk.ValAddress) error { return nil },
						}
						rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
						k.DeleteSigningSessionFunc = func(sdk.Context, uint64) {}
						sigHandler.HandleCompletedFunc = func(sdk.Context, utils.ValidatedProtoMarshaler, codec.ProtoMarshaler) error { return nil }

						_, err := multisig.EndBlocker(ctx, k, rewarder)

						assert.NoError(t, err)
						assert.Len(t, k.DeleteSigningSessionCalls(), 1)
						assert.Len(t, sigHandler.HandleCompletedCalls(), 1)
						assert.Len(t, pool.ReleaseRewardsCalls(), len(signingSession.Key.GetParticipants()))
					}),

				When("a completed signing session with missing participants and expiry equal to the block height", func() {
					missingCount = uint64(rand.I64Between(1, 5))
					signingSession = newSigningSessionWithMissingParticipants(module, missingCount)
					k.GetSigningSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.SigningSession {
						if expiry != ctx.BlockHeight()+1 {
							return nil
						}

						return []types.SigningSession{signingSession}
					}
				}).
					Then("should delete and set sig", func(t *testing.T) {
						pool := rewardmock.RewardPoolMock{
							ClearRewardsFunc:   func(sdk.ValAddress) {},
							ReleaseRewardsFunc: func(sdk.ValAddress) error { return nil },
						}
						rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
						k.DeleteSigningSessionFunc = func(sdk.Context, uint64) {}
						sigHandler.HandleCompletedFunc = func(sdk.Context, utils.ValidatedProtoMarshaler, codec.ProtoMarshaler) error { return nil }

						_, err := multisig.EndBlocker(ctx, k, rewarder)

						assert.NoError(t, err)
						assert.Len(t, k.DeleteSigningSessionCalls(), 1)
						assert.Len(t, sigHandler.HandleCompletedCalls(), 1)
						assert.Len(t, pool.ReleaseRewardsCalls(), len(signingSession.Key.GetParticipants())-int(missingCount))
						assert.Len(t, pool.ClearRewardsCalls(), int(missingCount))
					}),

				When("multiple completed signing sessions are triggered", func() {
					k.GetSigningSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.SigningSession {
						if expiry != ctx.BlockHeight()+1 {
							return nil
						}
						return []types.SigningSession{
							newSigningSessionWithMissingParticipants(module, uint64(rand.I64Between(1, 5))),
							newSigningSession(module),
							newSigningSessionWithMissingParticipants(module, uint64(rand.I64Between(1, 5))),
							newSigningSession(module),
						}
					}
				}).
					Then("should delete and set sig", func(t *testing.T) {
						pool := rewardmock.RewardPoolMock{
							ClearRewardsFunc:   func(sdk.ValAddress) {},
							ReleaseRewardsFunc: func(sdk.ValAddress) error { return nil },
						}
						rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
						k.DeleteSigningSessionFunc = func(sdk.Context, uint64) {}
						sigHandler.HandleCompletedFunc = func(sdk.Context, utils.ValidatedProtoMarshaler, codec.ProtoMarshaler) error { return nil }

						_, err := multisig.EndBlocker(ctx, k, rewarder)

						assert.NoError(t, err)
						assert.Len(t, k.DeleteSigningSessionCalls(), 4)
						assert.Len(t, sigHandler.HandleCompletedCalls(), 4)
					}),

				When("multiple completed signing sessions are triggered", func() {
					k.GetSigningSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.SigningSession {
						if expiry != ctx.BlockHeight()+1 {
							return nil
						}
						return []types.SigningSession{
							newSigningSessionWithMissingParticipants(module, uint64(rand.I64Between(1, 5))),
							newSigningSession(module),
							newSigningSessionWithMissingParticipants(module, uint64(rand.I64Between(1, 5))),
							newSigningSession(module),
						}
					}
				}).
					When("sigHandler fails", func() {
						sigHandler.HandleCompletedFunc = func(sdk.Context, utils.ValidatedProtoMarshaler, codec.ProtoMarshaler) error {
							return fmt.Errorf("some error")
						}
						sigHandler.HandleFailedFunc = func(sdk.Context, codec.ProtoMarshaler) error { return nil }
					}).
					Then("keep session cleanup and abort the signing", func(t *testing.T) {
						storeKey := store.NewKVStoreKey("cache")
						deletedKey := []byte("deleted")
						clearedKey := []byte("rewards-cleared")
						// the pool writes through the ctx it was requested with, so a
						// ClearRewards call on a rolled-back cached ctx leaves no trace
						rewarder.GetPoolFunc = func(c sdk.Context, _ string) reward.RewardPool {
							return &rewardmock.RewardPoolMock{
								ClearRewardsFunc:   func(sdk.ValAddress) { c.MultiStore().GetKVStore(storeKey).Set(clearedKey, []byte{}) },
								ReleaseRewardsFunc: func(sdk.ValAddress) error { return nil },
							}
						}
						k.DeleteSigningSessionFunc = func(ctx sdk.Context, _ uint64) {
							ctx.MultiStore().GetKVStore(storeKey).Set(deletedKey, []byte{})
						}
						_, err := multisig.EndBlocker(ctx, k, rewarder)
						assert.NoError(t, err)
						assert.True(t, ctx.MultiStore().GetKVStore(storeKey).Has(deletedKey))
						assert.True(t, ctx.MultiStore().GetKVStore(storeKey).Has(clearedKey))
						assert.Equal(t, len(k.DeleteSigningSessionCalls()), len(k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()+1)))
						assert.Equal(t, len(sigHandler.HandleFailedCalls()), len(k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()+1)))
					}),

				When("multiple completed signing sessions are triggered", func() {
					k.GetSigningSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.SigningSession {
						if expiry != ctx.BlockHeight()+1 {
							return nil
						}
						return []types.SigningSession{
							newSigningSession(module),
							newSigningSession(module),
						}
					}
				}).
					When("sigHandler panics", func() {
						sigHandler.HandleCompletedFunc = func(sdk.Context, utils.ValidatedProtoMarshaler, codec.ProtoMarshaler) error {
							panic("panic in sig handler")
						}
						sigHandler.HandleFailedFunc = func(sdk.Context, codec.ProtoMarshaler) error { return nil }
					}).
					Then("recover, keep session cleanup and abort the signing", func(t *testing.T) {
						pool := rewardmock.RewardPoolMock{
							ClearRewardsFunc:   func(sdk.ValAddress) {},
							ReleaseRewardsFunc: func(sdk.ValAddress) error { return nil },
						}
						rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }
						storeKey := store.NewKVStoreKey("cache")
						deletedKey := []byte("deleted")
						k.DeleteSigningSessionFunc = func(ctx sdk.Context, _ uint64) {
							ctx.MultiStore().GetKVStore(storeKey).Set(deletedKey, []byte{})
						}

						assert.NotPanics(t, func() {
							_, err := multisig.EndBlocker(ctx, k, rewarder)
							assert.NoError(t, err)
						})
						assert.True(t, ctx.MultiStore().GetKVStore(storeKey).Has(deletedKey))
						assert.Len(t, k.DeleteSigningSessionCalls(), 2)
						assert.Len(t, sigHandler.HandleFailedCalls(), 2)
					}),

				When("a pending signing session expiry equal to the block height", func() {
					k.GetSigningSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.SigningSession {
						if expiry != ctx.BlockHeight()+1 {
							return nil
						}

						return []types.SigningSession{{
							ID:     uint64(rand.PosI64()),
							Module: module,
							Key:    typestestutils.Key(),
							State:  exported.Pending,
						}}
					}
				}).
					When("HandleFailed panics", func() {
						sigHandler.HandleFailedFunc = func(sdk.Context, codec.ProtoMarshaler) error {
							panic("panic in HandleFailed")
						}
					}).
					Then("recover, keep session cleanup, forfeited rewards and expiry event", func(t *testing.T) {
						storeKey := store.NewKVStoreKey("cache")
						deletedKey := []byte("deleted")
						clearedKey := []byte("rewards-cleared")
						rewarder.GetPoolFunc = func(c sdk.Context, _ string) reward.RewardPool {
							return &rewardmock.RewardPoolMock{
								ClearRewardsFunc: func(sdk.ValAddress) { c.MultiStore().GetKVStore(storeKey).Set(clearedKey, []byte{}) },
							}
						}
						k.DeleteSigningSessionFunc = func(ctx sdk.Context, _ uint64) {
							ctx.MultiStore().GetKVStore(storeKey).Set(deletedKey, []byte{})
						}

						assert.NotPanics(t, func() {
							_, err := multisig.EndBlocker(ctx, k, rewarder)
							assert.NoError(t, err)
						})
						assert.True(t, ctx.MultiStore().GetKVStore(storeKey).Has(deletedKey))
						assert.True(t, ctx.MultiStore().GetKVStore(storeKey).Has(clearedKey))
						assert.True(t, hasEvent(ctx, &types.SigningExpired{}))
						assert.Len(t, k.DeleteSigningSessionCalls(), 1)
					}),
			).
			Run(t, 20)
	})
}

func hasEvent(ctx sdk.Context, event gogoproto.Message) bool {
	eventType := gogoproto.MessageName(event)
	for _, e := range ctx.EventManager().Events() {
		if e.Type == eventType {
			return true
		}
	}

	return false
}

func newSigningSession(module string) types.SigningSession {
	return newSigningSessionWithMissingParticipants(module, 0)
}

func newSigningSessionWithMissingParticipants(module string, missingCount uint64) types.SigningSession {
	sig := typestestutils.MultiSig()
	validators := maps.Keys(sig.GetSigs())
	validators = append(validators, slices.Expand(func(_ int) string { return rand.ValAddr().String() }, int(missingCount))...)

	pubKeys := make(map[string]exported.PublicKey)
	for _, v := range validators {
		pubKeys[v] = funcs.Must(btcec.NewPrivateKey()).PubKey().SerializeCompressed()
	}

	participants := make(map[string]snapshot.Participant)
	for _, v := range validators {
		participants[v] = snapshot.NewParticipant(funcs.Must(sdk.ValAddressFromBech32(v)), math.OneUint())
	}

	return types.SigningSession{
		MultiSig: sig,
		Key: types.Key{
			ID:      testutils.KeyID(),
			PubKeys: pubKeys,
			Snapshot: snapshot.Snapshot{
				Participants: participants,
				BondedWeight: math.OneUint().MulUint64(uint64(len(participants))),
			},
			SigningThreshold: utils.NewThreshold(int64(len(participants))-int64(missingCount), int64(len(participants))),
		},
		State:  exported.Completed,
		Module: module,
	}
}
