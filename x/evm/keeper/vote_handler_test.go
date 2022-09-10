package keeper_test

import (
	mathRand "math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	fakeMock "github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	mock4 "github.com/axelarnetwork/axelar-core/x/nexus/exported/mock"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	mock3 "github.com/axelarnetwork/axelar-core/x/reward/exported/mock"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	mock2 "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestHandleExpiredPoll(t *testing.T) {
	missingVoter := rand.ValAddr()
	var (
		ctx             sdk.Context
		n               *mock.NexusMock
		rewardPool      *mock3.RewardPoolMock
		poll            *mock2.PollMock
		maintainerState *mock4.MaintainerStateMock
		handler         vote.VoteHandler
	)

	givenVoteHandler := Given("the vote handler", func() {
		ctx = sdk.NewContext(&fakeMock.MultiStoreMock{}, tmproto.Header{}, false, log.TestingLogger())
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
		rewardPool = &mock3.RewardPoolMock{}
		r := &mock.RewarderMock{
			GetPoolFunc: func(sdk.Context, string) reward.RewardPool { return rewardPool },
		}
		handler = keeper.NewVoteHandler(encCfg.Codec, k, n, r)
	})

	givenVoteHandler.
		When("some voter failed to vote for poll", func() {
			poll = &mock2.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return append(slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10), missingVoter)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return !address.Equals(missingVoter) },
			}
		}).
		When("maintainer state can be found", func() {
			maintainerState = &mock4.MaintainerStateMock{}
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
			poll = &mock2.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return append(slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10), missingVoter)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return !address.Equals(missingVoter) },
			}
		}).
		When("maintainer state can not be found", func() {
			maintainerState = &mock4.MaintainerStateMock{}
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
		When("no voter failed to vote for poll", func() {
			poll = &mock2.PollMock{
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.I64Between(10, 100)) },
				GetRewardPoolNameFunc: func() (string, bool) { return rand.NormalizedStr(3), true },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: exported.Ethereum.Name}, true },
				GetVotersFunc: func() []sdk.ValAddress {
					return slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10)
				},
				HasVotedFunc: func(address sdk.ValAddress) bool { return true },
			}
		}).
		When("maintainer state can be found", func() {
			maintainerState = &mock4.MaintainerStateMock{}
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
		n       *mock.NexusMock
		r       *mock.RewarderMock
		result  codec.ProtoMarshaler
		handler vote.VoteHandler
	)

	setup := func() {
		ctx = sdk.NewContext(&fakeMock.MultiStoreMock{}, tmproto.Header{}, false, log.TestingLogger())

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				if chain.Equals(evmChain) {
					return chaink
				}
				return nil
			},
			LoggerFunc:   func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
			HasChainFunc: func(ctx sdk.Context, chain nexus.ChainName) bool { return true },
		}
		chaink = &mock.ChainKeeperMock{
			GetEventFunc: func(sdk.Context, types.EventID) (types.Event, bool) {
				return types.Event{}, false
			},
			SetConfirmedEventFunc: func(sdk.Context, types.Event) error {
				return nil
			},
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}

		chains := map[nexus.ChainName]nexus.Chain{
			exported.Ethereum.Name: exported.Ethereum,
		}
		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}
		r = &mock.RewarderMock{}
		encCfg := params.MakeEncodingConfig()
		handler = keeper.NewVoteHandler(encCfg.Codec, basek, n, r)
	}

	repeats := 20

	t.Run("Given vote When events are not from the same source chain THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		result = &types.VoteEvents{
			Chain:  nexus.ChainName(rand.Str(5)),
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("Given vote When events empty THEN should nothing and return nil", testutils.Func(func(t *testing.T) {
		setup()

		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: []types.Event{},
		}
		err := handler.HandleResult(ctx, result)

		assert.NoError(t, err)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not registered THEN return error", testutils.Func(func(t *testing.T) {
		setup()
		n.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{}, false
		}
		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN chain is not activated THEN still confirm the event", testutils.Func(func(t *testing.T) {
		setup()

		n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }

		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.NoError(t, err)
	}).Repeat(repeats))

	t.Run("GIVEN vote WHEN result is invalid THEN panic", testutils.Func(func(t *testing.T) {
		setup()

		result = types.NewConfirmGatewayTxRequest(rand.AccAddr(), rand.Str(5), types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))))
		assert.Panics(t, func() {
			_ = handler.HandleResult(ctx, result)
		})
	}).Repeat(repeats))

	t.Run("GIVEN already confirmed event WHEN handle deposit THEN return error", testutils.Func(func(t *testing.T) {
		setup()

		chaink.GetEventFunc = func(sdk.Context, types.EventID) (types.Event, bool) {
			return types.Event{}, true
		}

		result = &types.VoteEvents{
			Chain:  evmChain,
			Events: randTransferEvents(int(rand.I64Between(5, 10))),
		}
		err := handler.HandleResult(ctx, result)

		assert.Error(t, err)
	}).Repeat(repeats))

	var (
		poll      *mock2.PollMock
		nexusMock *mock.NexusMock
	)

	Given("a vote handler", func() {
		encCfg := params.MakeEncodingConfig()
		nexusMock = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					SupportsForeignAssets: true,
					Module:                types.ModuleName,
				}, true
			},
		}
		rewarder := &mock.RewarderMock{
			GetPoolFunc: func(sdk.Context, string) reward.RewardPool { return &mock3.RewardPoolMock{} },
		}
		handler = keeper.NewVoteHandler(encCfg.Codec, &mock.BaseKeeperMock{}, nexusMock, rewarder)
	}).
		Given("a completed poll", func() {
			poll = &mock2.PollMock{
				GetStateFunc:          func() vote.PollState { return vote.Completed },
				GetResultFunc:         func() codec.ProtoMarshaler { return &types.VoteEvents{Chain: "ethereum", Events: nil} },
				GetRewardPoolNameFunc: func() (string, bool) { return "rewards", true },
				GetIDFunc:             func() vote.PollID { return vote.PollID(rand.PosI64()) },
				GetMetaDataFunc:       func() (codec.ProtoMarshaler, bool) { return &types.PollMetadata{Chain: "ethereum"}, true },
				GetVotersFunc:         func() []sdk.ValAddress { return slices.Expand(func(int) sdk.ValAddress { return rand.ValAddr() }, 10) },
			}
		}).
		When("a voter is not a chain maintainer", func() {
			nexusMock.GetChainMaintainerStateFunc = func(sdk.Context, nexus.Chain, sdk.ValAddress) (nexus.MaintainerState, bool) {
				return nil, false
			}
		}).
		Then("ignore that voter", func(t *testing.T) {
			ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
			assert.NoError(t, handler.HandleCompletedPoll(ctx, poll))
			assert.Len(t, nexusMock.SetChainMaintainerStateCalls(), 0)
		}).Run(t)
}

func randTransferEvents(n int) []types.Event {
	events := make([]types.Event, n)
	burnerAddress := types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
	for i := 0; i < n; i++ {
		transfer := types.EventTransfer{
			To:     burnerAddress,
			Amount: sdk.NewUint(mathRand.Uint64()),
		}
		events[i] = types.Event{
			Chain: evmChain,
			TxID:  types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Index: uint64(rand.I64Between(1, 50)),
			Event: &types.Event_Transfer{
				Transfer: &transfer,
			},
		}
	}

	return events
}
