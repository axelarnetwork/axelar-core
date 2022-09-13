package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestPoll(t *testing.T) {
	var (
		ctx         sdk.Context
		k           keeper.Keeper
		voters      [4]sdk.ValAddress
		pollBuilder exported.PollBuilder
		poll        exported.Poll
	)

	for i := 0; i < len(voters); i++ {
		voters[i] = rand.ValAddr()
	}
	participants := slices.Map(voters[:], func(v sdk.ValAddress) snapshot.Participant {
		return snapshot.NewParticipant(v, sdk.OneUint())
	})

	givenPollBuilder := Given("a poll builder", func() {
		snapshotter := mock.SnapshotterMock{}
		staking := mock.StakingKeeperMock{}
		rewarder := mock.RewarderMock{}

		ctx = sdk.NewContext(fake.NewMultiStore(), abci.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		encodingConfig := params.MakeEncodingConfig()
		types.RegisterLegacyAminoCodec(encodingConfig.Amino)
		types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
		encodingConfig.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
		subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "vote")

		k = keeper.NewKeeper(
			encodingConfig.Codec,
			sdk.NewKVStoreKey(types.StoreKey),
			subspace,
			&snapshotter,
			&staking,
			&rewarder,
		)
		k.SetParams(ctx, types.DefaultParams())
		module := rand.NormalizedStr(5)

		snapshot := snapshot.NewSnapshot(time.Now(), rand.I64Between(1, 100), participants, sdk.NewUint(5))
		pollBuilder = exported.NewPollBuilder(
			module,
			utils.NewThreshold(51, 100),
			snapshot,
			ctx.BlockHeight()+100,
		).
			GracePeriod(1)
	})

	whenPollIsInitialized := When("poll is initialized", func() {
		pollID, err := k.InitializePoll(ctx, pollBuilder)
		if err != nil {
			panic(err)
		}

		poll, _ = k.GetPoll(ctx, pollID)
	})

	t.Run("HasVotedCorrectly", func(t *testing.T) {
		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should return whether or not the given voter has voted correctly", func(t *testing.T) {
				for _, voter := range voters {
					assert.False(t, poll.HasVotedCorrectly(voter))
				}

				for _, voter := range voters[0:3] {
					assert.Nil(t, poll.GetResult())
					poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}})
				}
				poll.Vote(voters[3], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{}})

				for _, voter := range voters[0:3] {
					assert.True(t, poll.HasVotedCorrectly(voter))
				}
				assert.False(t, poll.HasVotedCorrectly(voters[3]))
			}).
			Run(t)
	})

	t.Run("HasVoted", func(t *testing.T) {
		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should return whether or not the given voter has voted", func(t *testing.T) {
				for _, voter := range voters {
					assert.False(t, poll.HasVoted(voter))
				}

				for _, voter := range voters[0:3] {
					assert.Nil(t, poll.GetResult())
					poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}})
				}

				for _, voter := range voters[0:3] {
					assert.True(t, poll.HasVoted(voter))
				}
				assert.False(t, poll.HasVoted(voters[3]))
			}).
			Run(t)
	})

	t.Run("GetResult", func(t *testing.T) {
		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should return the correct result", func(t *testing.T) {
				expected := &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}}

				for _, voter := range voters[0:3] {
					assert.Nil(t, poll.GetResult())
					poll.Vote(voter, ctx.BlockHeight(), expected)
				}

				assert.NotNil(t, poll.GetResult())
				assert.Equal(t, poll.GetResult(), expected)
			}).
			Run(t)
	})

	t.Run("GetVoters", func(t *testing.T) {
		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should return all the voters", func(t *testing.T) {
				actual := poll.GetVoters()

				assert.ElementsMatch(t, voters, actual)
			}).
			Run(t)
	})

	t.Run("Vote", func(t *testing.T) {
		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should be able to vote for a pending poll and complete it", func(t *testing.T) {
				for _, voter := range voters[0:3] {
					assert.EqualValues(t, exported.Pending, poll.GetState())
					poll, _ = k.GetPoll(ctx, poll.GetID())
					assert.EqualValues(t, exported.Pending, poll.GetState())

					voteResult, err := poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{})

					assert.NoError(t, err)
					assert.EqualValues(t, exported.VoteInTime, voteResult)
				}

				assert.EqualValues(t, exported.Completed, poll.GetState())
				poll, _ = k.GetPoll(ctx, poll.GetID())
				assert.EqualValues(t, exported.Completed, poll.GetState())
			}).
			Run(t)

		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should be able to complete multiple polls in a row", func(t *testing.T) {
				originalPollID := poll.GetID()

				for _, voter := range voters {
					poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{})
				}

				assert.EqualValues(t, exported.Completed, poll.GetState())

				module := rand.NormalizedStr(5)
				snapshot := snapshot.NewSnapshot(time.Now(), rand.I64Between(1, 100), participants, sdk.NewUint(5))
				pollBuilder = exported.NewPollBuilder(
					module,
					utils.NewThreshold(51, 100),
					snapshot,
					ctx.BlockHeight()+100,
				).
					GracePeriod(1)
				pollID, err := k.InitializePoll(ctx, pollBuilder)
				if err != nil {
					panic(err)
				}
				assert.NotEqual(t, originalPollID, pollID)
				poll, _ = k.GetPoll(ctx, pollID)

				for _, voter := range voters {
					poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{})
				}

				assert.EqualValues(t, exported.Completed, poll.GetState())
			}).
			Run(t)

		givenPollBuilder.
			When("min voter count is set", func() { pollBuilder = pollBuilder.MinVoterCount(int64(len(voters))) }).
			When2(whenPollIsInitialized).
			Then("should only complete the poll when min voter count is hit", func(t *testing.T) {
				for _, voter := range voters {
					assert.EqualValues(t, exported.Pending, poll.GetState())
					poll, _ = k.GetPoll(ctx, poll.GetID())
					assert.EqualValues(t, exported.Pending, poll.GetState())

					voteResult, err := poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{})

					assert.NoError(t, err)
					assert.EqualValues(t, exported.VoteInTime, voteResult)
				}

				assert.EqualValues(t, exported.Completed, poll.GetState())
				poll, _ = k.GetPoll(ctx, poll.GetID())
				assert.EqualValues(t, exported.Completed, poll.GetState())
			}).
			Run(t)

		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should be able to vote for a completed poll within the grace period", func(t *testing.T) {
				for _, voter := range voters[0:3] {
					poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{})
				}

				voteResult, err := poll.Vote(voters[3], ctx.BlockHeight()+1, &evmtypes.VoteEvents{})

				assert.NoError(t, err)
				assert.EqualValues(t, exported.VotedLate, voteResult)
				assert.EqualValues(t, exported.Completed, poll.GetState())
			}).
			Run(t)

		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should not be able to vote for a completed poll outside the grace period", func(t *testing.T) {
				for _, voter := range voters[0:3] {
					poll.Vote(voter, ctx.BlockHeight(), &evmtypes.VoteEvents{})
				}

				voteResult, err := poll.Vote(voters[3], ctx.BlockHeight()+2, &evmtypes.VoteEvents{})

				assert.NoError(t, err)
				assert.EqualValues(t, exported.NoVote, voteResult)
				assert.EqualValues(t, exported.Completed, poll.GetState())
			}).
			Run(t)

		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should not be able to re-vote", func(t *testing.T) {
				poll.Vote(voters[0], ctx.BlockHeight(), &evmtypes.VoteEvents{})
				voteResult, err := poll.Vote(voters[0], ctx.BlockHeight(), &evmtypes.VoteEvents{})

				assert.Error(t, err)
				assert.EqualValues(t, exported.NoVote, voteResult)
			}).
			Run(t)

		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should not allow non-voters to vote", func(t *testing.T) {
				voteResult, err := poll.Vote(rand.ValAddr(), ctx.BlockHeight(), &evmtypes.VoteEvents{})

				assert.Error(t, err)
				assert.EqualValues(t, exported.NoVote, voteResult)
			}).
			Run(t)

		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should fail the poll if it is impossible to pass the threshold", func(t *testing.T) {
				poll.Vote(voters[0], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}})
				poll.Vote(voters[1], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}, {}}})
				voteResult, err := poll.Vote(voters[2], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}, {}, {}}})

				assert.NoError(t, err)
				assert.EqualValues(t, exported.VoteInTime, voteResult)

				assert.EqualValues(t, exported.Failed, poll.GetState())
				poll, _ = k.GetPoll(ctx, poll.GetID())
				assert.EqualValues(t, exported.Failed, poll.GetState())
			}).
			Run(t)

		givenPollBuilder.
			When2(whenPollIsInitialized).
			Then("should not be able to vote for a failed poll", func(t *testing.T) {
				poll.Vote(voters[0], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}})
				poll.Vote(voters[1], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}, {}}})
				poll.Vote(voters[2], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}, {}, {}}})

				voteResult, err := poll.Vote(voters[3], ctx.BlockHeight(), &evmtypes.VoteEvents{Events: []evmtypes.Event{{}, {}, {}}})

				assert.NoError(t, err)
				assert.EqualValues(t, exported.NoVote, voteResult)
				assert.EqualValues(t, exported.Failed, poll.GetState())
			}).
			Run(t)
	})
}

func TestPoll_GetMetaData(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	evmtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "vote")
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.RewarderMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	snap := snapshot.NewSnapshot(
		time.Now(),
		rand.I64Between(1, 100),
		slices.Expand(func(_ int) snapshot.Participant { return snapshot.NewParticipant(rand.ValAddr(), sdk.OneUint()) }, 5),
		sdk.NewUint(5),
	)
	expectedMetadata := &evmtypes.PollMetadata{
		Chain: "chain",
		TxID:  [common.HashLength]byte{},
	}
	pollBuilder := exported.NewPollBuilder(
		"some_module",
		utils.NewThreshold(51, 100),
		snap,
		ctx.BlockHeight()+100,
	).
		GracePeriod(1).
		ModuleMetadata(expectedMetadata)

	pollID := funcs.Must(k.InitializePoll(ctx, pollBuilder))

	poll := funcs.MustOk(k.GetPoll(ctx, pollID))

	md, ok := poll.GetMetaData()
	assert.True(t, ok)
	assert.Equal(t, expectedMetadata, md)

}
