package keeper_test

import (
	"fmt"
	mathRand "math/rand"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkstore "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	fakemock "github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusmock "github.com/axelarnetwork/axelar-core/x/nexus/exported/mock"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	rewardmock "github.com/axelarnetwork/axelar-core/x/reward/exported/mock"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	votemock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestHandleExpiredPoll(t *testing.T) {
	missingVoter := rand.ValAddr()
	var (
		ctx             sdk.Context
		n               *mock.NexusMock
		rewardPool      *rewardmock.RewardPoolMock
		poll            *votemock.PollMock
		maintainerState *nexusmock.MaintainerStateMock
		handler         vote.VoteHandler
	)

	givenVoteHandler := Given("the vote handler", func() {
		ctx = sdk.NewContext(&fakemock.MultiStoreMock{}, tmproto.Header{}, false, log.TestingLogger())
		encCfg := params.MakeEncodingConfig()

		k := &mock.BaseKeeperMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}
		n = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					SupportsForeignAssets: true,
					Module:                types.ModuleName,
				}, true
			},
		}
		rewardPool = &rewardmock.RewardPoolMock{}
		r := &mock.RewarderMock{
			GetPoolFunc: func(sdk.Context, string) reward.RewardPool { return rewardPool },
		}
		handler = keeper.NewVoteHandler(encCfg.Codec, k, n, r)
	})

	givenVoteHandler.
		When("some voter failed to vote for poll", func() {
			poll = &votemock.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return append(slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10), missingVoter)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return !address.Equals(missingVoter) },
			}
		}).
		When("the voter is a chain maintainer", func() {
			maintainerState = &nexusmock.MaintainerStateMock{}
			n.GetChainMaintainerStateFunc = func(sdk.Context, nexus.Chain, sdk.ValAddress) (nexus.MaintainerState, bool) {
				return maintainerState, true
			}
		}).
		Then("should clear rewards and mark voter missing vote", func(t *testing.T) {
			maintainerState.MarkMissingVoteFunc = func(bool) {}
			n.SetChainMaintainerStateFunc = func(ctx sdk.Context, maintainerState nexus.MaintainerState) error { return nil }
			rewardPool.ClearRewardsFunc = func(sdk.ValAddress) {}

			err := handler.HandleExpiredPoll(ctx, poll)

			assert.NoError(t, err)
			assert.Len(t, maintainerState.MarkMissingVoteCalls(), 11)
			for i, call := range maintainerState.MarkMissingVoteCalls() {
				assert.Equal(t, i == len(maintainerState.MarkMissingVoteCalls())-1, call.MissingVote)
			}
			assert.Len(t, n.SetChainMaintainerStateCalls(), 11)
			assert.Len(t, rewardPool.ClearRewardsCalls(), 1)
			assert.Equal(t, missingVoter, rewardPool.ClearRewardsCalls()[0].ValAddress)
		}).
		Run(t)

	givenVoteHandler.
		When("some voter failed to vote for poll", func() {
			poll = &votemock.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return append(slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10), missingVoter)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return !address.Equals(missingVoter) },
			}
		}).
		When("the voter is not a chain maintainer", func() {
			maintainerState = &nexusmock.MaintainerStateMock{}
			n.GetChainMaintainerStateFunc = func(sdk.Context, nexus.Chain, sdk.ValAddress) (nexus.MaintainerState, bool) {
				return nil, false
			}
		}).
		Then("should clear rewards and not mark voter missing vote", func(t *testing.T) {
			rewardPool.ClearRewardsFunc = func(sdk.ValAddress) {}

			err := handler.HandleExpiredPoll(ctx, poll)

			assert.NoError(t, err)
			assert.Len(t, rewardPool.ClearRewardsCalls(), 1)
			assert.Equal(t, missingVoter, rewardPool.ClearRewardsCalls()[0].ValAddress)
		}).
		Run(t)

	givenVoteHandler.
		When("all voters failed to vote for poll", func() {
			poll = &votemock.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return false },
			}
		}).
		When("the voters are a chain maintainer", func() {
			maintainerState = &nexusmock.MaintainerStateMock{}
			n.GetChainMaintainerStateFunc = func(sdk.Context, nexus.Chain, sdk.ValAddress) (nexus.MaintainerState, bool) {
				return maintainerState, true
			}
		}).
		Then("should clear rewards and mark voters missing vote", func(t *testing.T) {
			maintainerState.MarkMissingVoteFunc = func(bool) {}
			n.SetChainMaintainerStateFunc = func(ctx sdk.Context, maintainerState nexus.MaintainerState) error { return nil }
			rewardPool.ClearRewardsFunc = func(sdk.ValAddress) {}

			err := handler.HandleExpiredPoll(ctx, poll)

			assert.NoError(t, err)
			assert.Len(t, maintainerState.MarkMissingVoteCalls(), 10)
			for _, call := range maintainerState.MarkMissingVoteCalls() {
				assert.True(t, call.MissingVote)
			}
			assert.Len(t, n.SetChainMaintainerStateCalls(), 10)
			assert.Len(t, rewardPool.ClearRewardsCalls(), 10)
		}).
		Run(t)

	givenVoteHandler.
		When("all voters failed to vote for poll", func() {
			poll = &votemock.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return false },
			}
		}).
		When("the voters are not a chain maintainer", func() {
			maintainerState = &nexusmock.MaintainerStateMock{}
			n.GetChainMaintainerStateFunc = func(sdk.Context, nexus.Chain, sdk.ValAddress) (nexus.MaintainerState, bool) {
				return nil, false
			}
		}).
		Then("should clear rewards and not mark voters missing vote", func(t *testing.T) {
			rewardPool.ClearRewardsFunc = func(sdk.ValAddress) {}

			err := handler.HandleExpiredPoll(ctx, poll)

			assert.NoError(t, err)
			assert.Len(t, rewardPool.ClearRewardsCalls(), 10)
		}).
		Run(t)

	givenVoteHandler.
		When("no voter failed to vote for poll", func() {
			poll = &votemock.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return true },
			}
		}).
		When("the voter is a chain maintainer", func() {
			maintainerState = &nexusmock.MaintainerStateMock{}
			n.GetChainMaintainerStateFunc = func(sdk.Context, nexus.Chain, sdk.ValAddress) (nexus.MaintainerState, bool) {
				return maintainerState, true
			}
		}).
		Then("should not clear rewards and not mark voter missing vote", func(t *testing.T) {
			maintainerState.MarkMissingVoteFunc = func(bool) {}
			n.SetChainMaintainerStateFunc = func(ctx sdk.Context, maintainerState nexus.MaintainerState) error { return nil }

			err := handler.HandleExpiredPoll(ctx, poll)

			assert.NoError(t, err)
			assert.Len(t, maintainerState.MarkMissingVoteCalls(), 10)
			for _, call := range maintainerState.MarkMissingVoteCalls() {
				assert.False(t, call.MissingVote)
			}
			assert.Len(t, n.SetChainMaintainerStateCalls(), 10)
		}).
		Run(t)
}

func TestHandleResult(t *testing.T) {
	var (
		ctx     sdk.Context
		basek   *mock.BaseKeeperMock
		chaink  *mock.ChainKeeperMock
		nexusK  *mock.NexusMock
		result  codec.ProtoMarshaler
		handler vote.VoteHandler
	)

	setup := func() {
		multiStore := fakemock.MultiStoreMock{}
		multiStore.CacheMultiStoreFunc = func() sdkstore.CacheMultiStore { return &fakemock.CacheMultiStoreMock{} }
		ctx = sdk.NewContext(&multiStore, tmproto.Header{}, false, log.TestingLogger())

		basek = &mock.BaseKeeperMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}
		chaink = &mock.ChainKeeperMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}
		nexusK = &mock.NexusMock{}

		handler = keeper.NewVoteHandler(params.MakeEncodingConfig().Codec, basek, nexusK, &mock.RewarderMock{})
	}

	givenHandler := Given("the vote handler", setup)

	givenHandler.
		When("result is falsy", func() {
			chain := nexus.ChainName(rand.Str(5))
			result = &types.VoteEvents{
				Chain:  chain,
				Events: nil,
			}
		}).
		Then("should return nil and do nothing", func(t *testing.T) {
			assert.NoError(t, handler.HandleResult(ctx, result))
		}).
		Run(t)

	givenHandler.
		When("source chain is not registered", func() {
			chain := nexus.ChainName(rand.Str(5))
			result = &types.VoteEvents{
				Chain:  chain,
				Events: randTransferEvents(chain, rand.I64Between(5, 10)),
			}

			nexusK.GetChainFunc = func(_ sdk.Context, _ nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, false }
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, handler.HandleResult(ctx, result), "is not a registered chain")
		}).
		Run(t)

	givenHandler.
		When("source chain is not an evm chain", func() {
			chain := nexus.ChainName(rand.Str(5))
			result = &types.VoteEvents{
				Chain:  chain,
				Events: randTransferEvents(chain, rand.I64Between(5, 10)),
			}

			nexusK.GetChainFunc = func(_ sdk.Context, _ nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, true }
			basek.ForChainFunc = func(_ sdk.Context, _ nexus.ChainName) (types.ChainKeeper, error) {
				return nil, fmt.Errorf("not an evm chain")
			}
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, handler.HandleResult(ctx, result), "is not an evm chain")
		}).
		Run(t)

	givenHandler.
		When("source chain is an evm chain", func() {
			result = &types.VoteEvents{
				Chain: exported.Ethereum.Name,
			}

			nexusK.GetChainFunc = func(_ sdk.Context, chainName nexus.ChainName) (nexus.Chain, bool) {
				switch chainName {
				case exported.Ethereum.Name:
					return exported.Ethereum, true
				case axelarnet.Axelarnet.Name:
					return axelarnet.Axelarnet, true
				default:
					return nexus.Chain{}, false
				}
			}
			basek.ForChainFunc = func(_ sdk.Context, _ nexus.ChainName) (types.ChainKeeper, error) {
				return chaink, nil
			}
		}).
		Branch(
			When("failed to set the confirmed event", func() {
				result.(*types.VoteEvents).Events = randTransferEvents(exported.Ethereum.Name, 1)

				chaink.SetConfirmedEventFunc = func(_ sdk.Context, _ types.Event) error { return fmt.Errorf("failed to set confirmed event") }
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, handler.HandleResult(ctx, result), "failed to set confirmed event")
				}),

			When("event is not contract call", func() {
				result.(*types.VoteEvents).Events = randTransferEvents(exported.Ethereum.Name, 5)
			}).
				When("succeeded to set the confirmed event", func() {
					chaink.SetConfirmedEventFunc = func(_ sdk.Context, _ types.Event) error { return nil }
				}).
				Then("should enqueue the confirmed event", func(t *testing.T) {
					chaink.EnqueueConfirmedEventFunc = func(_ sdk.Context, _ types.EventID) error { return nil }

					assert.NoError(t, handler.HandleResult(ctx, result))
					assert.Len(t, chaink.EnqueueConfirmedEventCalls(), 5)
				}),

			When("event is contract call and is sent to a known chain", func() {
				result.(*types.VoteEvents).Events = randContractCallEvents(exported.Ethereum.Name, exported.Ethereum.Name, 5)
			}).
				When("succeeded to set the confirmed event", func() {
					chaink.SetConfirmedEventFunc = func(_ sdk.Context, _ types.Event) error { return nil }
				}).
				When("succeeded to route the general messages", func() {
					nexusK.EnqueueRouteMessageFunc = func(_ sdk.Context, _ string) error { return nil }
					chaink.SetEventCompletedFunc = func(sdk.Context, types.EventID) error { return nil }
				}).
				Then("should route the general messages", func(t *testing.T) {
					nexusK.SetNewMessageFunc = func(_ sdk.Context, _ nexus.GeneralMessage) error { return nil }

					assert.NoError(t, handler.HandleResult(ctx, result))
					assert.Len(t, nexusK.SetNewMessageCalls(), 5)
					assert.Len(t, nexusK.EnqueueRouteMessageCalls(), 5)
					assert.Len(t, chaink.SetEventCompletedCalls(), 5)
				}),

			When("event is contract call and is sent to a known chain", func() {
				result.(*types.VoteEvents).Events = randContractCallEvents(exported.Ethereum.Name, exported.Ethereum.Name, 5)
			}).
				When("succeeded to set the confirmed event", func() {
					chaink.SetConfirmedEventFunc = func(_ sdk.Context, _ types.Event) error { return nil }
				}).
				When("failed to route the general messages", func() {
					nexusK.EnqueueRouteMessageFunc = func(_ sdk.Context, _ string) error { return fmt.Errorf("failed") }
				}).
				Then("should panic", func(t *testing.T) {
					nexusK.SetNewMessageFunc = func(_ sdk.Context, _ nexus.GeneralMessage) error { return nil }

					assert.Panics(t, func() { handler.HandleResult(ctx, result) })
					assert.Len(t, nexusK.SetNewMessageCalls(), 1)
					assert.Len(t, nexusK.EnqueueRouteMessageCalls(), 1)
				}),

			When("event is contract call and is sent to an unknown chain", func() {
				result.(*types.VoteEvents).Events = randContractCallEvents(exported.Ethereum.Name, nexus.ChainName(rand.Str(5)), 5)
			}).
				When("succeeded to set the confirmed event", func() {
					chaink.SetConfirmedEventFunc = func(_ sdk.Context, _ types.Event) error { return nil }
				}).
				When("succeeded to route the general messages", func() {
					nexusK.EnqueueRouteMessageFunc = func(_ sdk.Context, _ string) error { return nil }
					chaink.SetEventCompletedFunc = func(sdk.Context, types.EventID) error { return nil }
				}).
				Then("should set as approved general messages", func(t *testing.T) {
					nexusK.SetNewMessageFunc = func(_ sdk.Context, _ nexus.GeneralMessage) error { return nil }

					assert.NoError(t, handler.HandleResult(ctx, result))
					assert.Len(t, nexusK.SetNewMessageCalls(), 5)
					assert.Len(t, nexusK.EnqueueRouteMessageCalls(), 5)
					assert.Len(t, chaink.SetEventCompletedCalls(), 5)

					for _, call := range nexusK.SetNewMessageCalls() {
						assert.Equal(t, wasm.ModuleName, call.M.Recipient.Chain.Module)
					}
				}),
		).
		Run(t)
}

func randTransferEvents(chain nexus.ChainName, n int64) []types.Event {
	events := make([]types.Event, n)
	burnerAddress := types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
	for i := int64(0); i < n; i++ {
		transfer := types.EventTransfer{
			To:     burnerAddress,
			Amount: sdk.NewUint(mathRand.Uint64()),
		}
		events[i] = types.Event{
			Chain: chain,
			TxID:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Index: uint64(rand.I64Between(1, 50)),
			Event: &types.Event_Transfer{
				Transfer: &transfer,
			},
		}
	}

	return events
}

func randContractCallEvents(chain nexus.ChainName, destinationChain nexus.ChainName, n int64) []types.Event {
	events := make([]types.Event, n)
	for i := int64(0); i < n; i++ {
		contractCall := types.EventContractCall{
			Sender:           types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			DestinationChain: destinationChain,
			ContractAddress:  common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
			PayloadHash:      types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		}
		events[i] = types.Event{
			Chain: chain,
			TxID:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Index: uint64(rand.I64Between(1, 50)),
			Event: &types.Event_ContractCall{
				ContractCall: &contractCall,
			},
		}
	}

	return events
}
