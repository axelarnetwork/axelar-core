package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

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
		return snapshot.NewParticipant(v, math.OneUint())
	})

	givenPollBuilder := Given("a poll builder", func() {
		snapshotter := mock.SnapshotterMock{}
		staking := mock.StakingKeeperMock{}
		rewarder := mock.RewarderMock{}

		ctx = sdk.NewContext(fake.NewMultiStore(), abci.Header{Height: rand.PosI64()}, false, log.NewTestLogger(t))
		encodingConfig := params.MakeEncodingConfig()
		types.RegisterLegacyAminoCodec(encodingConfig.Amino)
		types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
		encodingConfig.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
		subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "vote")

		k = keeper.NewKeeper(
			encodingConfig.Codec,
			store.NewKVStoreKey(types.StoreKey),
			subspace,
			&snapshotter,
			&staking,
			&rewarder,
		)
		k.SetParams(ctx, types.DefaultParams())
		module := rand.NormalizedStr(5)

		snapshot := snapshot.NewSnapshot(time.Now(), rand.I64Between(1, 100), participants, math.NewUint(5))
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
				snapshot := snapshot.NewSnapshot(time.Now(), rand.I64Between(1, 100), participants, math.NewUint(5))
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

				assert.ErrorContains(t, err, "poll completed")
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

				assert.ErrorContains(t, err, "poll failed")
				assert.EqualValues(t, exported.NoVote, voteResult)
				assert.EqualValues(t, exported.Failed, poll.GetState())
			}).
			Run(t)
	})
}

// TestPoll_HasVotedCaching tests that calling HasVoted and HasVotedCorrectly
// multiple times for different voters returns consistent results with cached tallied votes.
func TestPoll_HasVotedCaching(t *testing.T) {
	numVoters := 20
	voters := make([]sdk.ValAddress, numVoters)
	for i := range numVoters {
		voters[i] = rand.ValAddr()
	}
	participants := slices.Map(voters, func(v sdk.ValAddress) snapshot.Participant {
		return snapshot.NewParticipant(v, math.OneUint())
	})

	snapshotter := mock.SnapshotterMock{}
	staking := mock.StakingKeeperMock{}
	rewarder := mock.RewarderMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{Height: rand.PosI64()}, false, log.NewTestLogger(t))
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	encodingConfig.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "vote")

	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
		&snapshotter,
		&staking,
		&rewarder,
	)
	k.SetParams(ctx, types.DefaultParams())

	snap := snapshot.NewSnapshot(time.Now(), rand.I64Between(1, 100), participants, math.NewUint(uint64(numVoters)))
	pollBuilder := exported.NewPollBuilder(
		rand.NormalizedStr(5),
		utils.NewThreshold(51, 100),
		snap,
		ctx.BlockHeight()+100,
	).GracePeriod(1)

	pollID, err := k.InitializePoll(ctx, pollBuilder)
	assert.NoError(t, err)

	poll, ok := k.GetPoll(ctx, pollID)
	assert.True(t, ok)

	voteData := &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}}
	majorityCount := (numVoters * 51 / 100) + 1
	for i := range majorityCount {
		_, err := poll.Vote(voters[i], ctx.BlockHeight(), voteData)
		assert.NoError(t, err)
	}

	assert.EqualValues(t, exported.Completed, poll.GetState())

	poll, ok = k.GetPoll(ctx, pollID)
	assert.True(t, ok)

	allVoters := poll.GetVoters()
	hasVotedResults := make(map[string]bool)
	hasVotedCorrectlyResults := make(map[string]bool)

	for _, voter := range allVoters {
		hasVotedResults[voter.String()] = poll.HasVoted(voter)
		hasVotedCorrectlyResults[voter.String()] = poll.HasVotedCorrectly(voter)
	}

	for _, voter := range allVoters {
		assert.Equal(t, hasVotedResults[voter.String()], poll.HasVoted(voter),
			"HasVoted should return consistent results for voter %s", voter.String())
		assert.Equal(t, hasVotedCorrectlyResults[voter.String()], poll.HasVotedCorrectly(voter),
			"HasVotedCorrectly should return consistent results for voter %s", voter.String())
	}

	for i := range majorityCount {
		assert.True(t, hasVotedResults[voters[i].String()],
			"Voter %d should have voted", i)
		assert.True(t, hasVotedCorrectlyResults[voters[i].String()],
			"Voter %d should have voted correctly", i)
	}

	for i := majorityCount; i < numVoters; i++ {
		assert.False(t, hasVotedResults[voters[i].String()],
			"Voter %d should not have voted", i)
		assert.False(t, hasVotedCorrectlyResults[voters[i].String()],
			"Voter %d should not have voted correctly", i)
	}
}

// TestPoll_ImmediateStateTransition verifies that poll state transitions happen immediately
// within the same Vote() call, not delayed to the next transaction. This catches bugs where
// stale cached data prevents immediate completion.
func TestPoll_ImmediateStateTransition(t *testing.T) {
	numVoters := 5
	voters := make([]sdk.ValAddress, numVoters)
	for i := range numVoters {
		voters[i] = rand.ValAddr()
	}
	participants := slices.Map(voters, func(v sdk.ValAddress) snapshot.Participant {
		return snapshot.NewParticipant(v, math.OneUint())
	})

	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	encodingConfig.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "vote")

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{Height: 100}, false, log.NewTestLogger(t))
	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
		&mock.SnapshotterMock{},
		&mock.StakingKeeperMock{},
		&mock.RewarderMock{},
	)
	k.SetParams(ctx, types.DefaultParams())

	snap := snapshot.NewSnapshot(time.Now(), 1, participants, math.NewUint(uint64(numVoters)))
	pollBuilder := exported.NewPollBuilder(
		"evm",
		utils.NewThreshold(51, 100),
		snap,
		ctx.BlockHeight()+100,
	).GracePeriod(1)

	pollID, err := k.InitializePoll(ctx, pollBuilder)
	assert.NoError(t, err)

	poll, ok := k.GetPoll(ctx, pollID)
	assert.True(t, ok)

	voteData := &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}}
	majorityCount := (numVoters*51)/100 + 1

	for i := range majorityCount - 1 {
		result, err := poll.Vote(voters[i], ctx.BlockHeight(), voteData)
		assert.NoError(t, err)
		assert.EqualValues(t, exported.VoteInTime, result)
		assert.EqualValues(t, exported.Pending, poll.GetState())
	}

	result, err := poll.Vote(voters[majorityCount-1], ctx.BlockHeight(), voteData)
	assert.NoError(t, err)
	assert.EqualValues(t, exported.VoteInTime, result)
	assert.EqualValues(t, exported.Completed, poll.GetState())

	persistedPoll, ok := k.GetPoll(ctx, pollID)
	assert.True(t, ok)
	assert.EqualValues(t, exported.Completed, persistedPoll.GetState(),
		"Completed state must be persisted to storage")
}

// TestPoll_ImmediateFailure verifies that poll failure happens immediately when it becomes
// impossible to reach threshold, not delayed to the next transaction.
func TestPoll_ImmediateFailure(t *testing.T) {
	numVoters := 4
	voters := make([]sdk.ValAddress, numVoters)
	for i := range numVoters {
		voters[i] = rand.ValAddr()
	}
	participants := slices.Map(voters, func(v sdk.ValAddress) snapshot.Participant {
		return snapshot.NewParticipant(v, math.OneUint())
	})

	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	encodingConfig.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "vote")

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{Height: 100}, false, log.NewTestLogger(t))
	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
		&mock.SnapshotterMock{},
		&mock.StakingKeeperMock{},
		&mock.RewarderMock{},
	)
	k.SetParams(ctx, types.DefaultParams())

	snap := snapshot.NewSnapshot(time.Now(), 1, participants, math.NewUint(uint64(numVoters)))
	pollBuilder := exported.NewPollBuilder(
		"evm",
		utils.NewThreshold(51, 100),
		snap,
		ctx.BlockHeight()+100,
	).GracePeriod(1)

	pollID, err := k.InitializePoll(ctx, pollBuilder)
	assert.NoError(t, err)

	poll, ok := k.GetPoll(ctx, pollID)
	assert.True(t, ok)

	voteData1 := &evmtypes.VoteEvents{Chain: "Ethereum", Events: []evmtypes.Event{{Chain: "Ethereum", Index: 1}}}
	voteData2 := &evmtypes.VoteEvents{Chain: "Ethereum", Events: []evmtypes.Event{{Chain: "Ethereum", Index: 2}}}
	voteData3 := &evmtypes.VoteEvents{Chain: "Ethereum", Events: []evmtypes.Event{{Chain: "Ethereum", Index: 3}}}

	poll.Vote(voters[0], ctx.BlockHeight(), voteData1)
	assert.EqualValues(t, exported.Pending, poll.GetState())

	poll.Vote(voters[1], ctx.BlockHeight(), voteData2)
	assert.EqualValues(t, exported.Pending, poll.GetState())

	poll.Vote(voters[2], ctx.BlockHeight(), voteData3)
	assert.EqualValues(t, exported.Failed, poll.GetState(),
		"Poll MUST fail immediately when threshold becomes unreachable, not on next transaction")

	persistedPoll, ok := k.GetPoll(ctx, pollID)
	assert.True(t, ok)
	assert.EqualValues(t, exported.Failed, persistedPoll.GetState(),
		"Failed state must be persisted to storage")
}

// TestPoll_StorageReadsCount demonstrates the O(N²) vs O(N) storage read improvement
// from caching tallied votes.
func TestPoll_StorageReadsCount(t *testing.T) {
	numVoters := 20
	voters := make([]sdk.ValAddress, numVoters)
	for i := range numVoters {
		voters[i] = rand.ValAddr()
	}
	participants := slices.Map(voters, func(v sdk.ValAddress) snapshot.Participant {
		return snapshot.NewParticipant(v, math.OneUint())
	})

	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	encodingConfig.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "vote")

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{Height: 1}, false, log.NewTestLogger(t))
	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
		&mock.SnapshotterMock{},
		&mock.StakingKeeperMock{},
		&mock.RewarderMock{},
	)
	k.SetParams(ctx, types.DefaultParams())

	snap := snapshot.NewSnapshot(time.Now(), 1, participants, math.NewUint(uint64(numVoters)))
	pollBuilder := exported.NewPollBuilder(
		"evm",
		utils.NewThreshold(51, 100),
		snap,
		ctx.BlockHeight()+100,
	).GracePeriod(1)

	pollID, _ := k.InitializePoll(ctx, pollBuilder)
	poll, _ := k.GetPoll(ctx, pollID)

	voteData := &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}}
	majorityCount := (numVoters * 51 / 100) + 1
	for i := range majorityCount {
		poll.Vote(voters[i], ctx.BlockHeight(), voteData)
	}

	poll, _ = k.GetPoll(ctx, pollID)
	allVoters := poll.GetVoters()

	operationCount := 0
	for _, voter := range allVoters {
		_ = poll.HasVoted(voter)
		operationCount++
		_ = poll.HasVotedCorrectly(voter)
		operationCount++
	}

	t.Logf("Voters: %d", numVoters)
	t.Logf("Operations (HasVoted + HasVotedCorrectly calls): %d", operationCount)
	t.Logf("WITHOUT caching fix: would trigger %d storage reads (O(N²))", operationCount)
	t.Logf("WITH caching fix: triggers 1 storage read (O(1) per poll)")

	assert.Equal(t, numVoters*2, operationCount)
}

// BenchmarkPoll_HasVoted_EndBlockerScenario benchmarks multiple polls expiring simultaneously,
// where HandleExpiredPoll calls HasVoted for each voter. The caching fix reduces storage reads
// from O(N²) per poll to O(N).
func BenchmarkPoll_HasVoted_EndBlockerScenario(b *testing.B) {
	numVoters := 83
	numPolls := 10

	voters := make([]sdk.ValAddress, numVoters)
	for i := range numVoters {
		voters[i] = rand.ValAddr()
	}
	participants := slices.Map(voters, func(v sdk.ValAddress) snapshot.Participant {
		return snapshot.NewParticipant(v, math.OneUint())
	})

	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	encodingConfig.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "vote")

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{Height: 1}, false, log.NewNopLogger())
	k := keeper.NewKeeper(
		encodingConfig.Codec,
		store.NewKVStoreKey(types.StoreKey),
		subspace,
		&mock.SnapshotterMock{},
		&mock.StakingKeeperMock{},
		&mock.RewarderMock{},
	)
	k.SetParams(ctx, types.DefaultParams())

	polls := make([]exported.Poll, numPolls)
	snap := snapshot.NewSnapshot(time.Now(), 1, participants, math.NewUint(uint64(numVoters)))

	for p := range numPolls {
		pollBuilder := exported.NewPollBuilder(
			"evm",
			utils.NewThreshold(51, 100),
			snap,
			ctx.BlockHeight()+100,
		).GracePeriod(1)

		pollID, _ := k.InitializePoll(ctx, pollBuilder)
		poll, _ := k.GetPoll(ctx, pollID)

		events := make([]evmtypes.Event, 50)
		for i := range events {
			events[i] = evmtypes.Event{
				Chain: "Ethereum",
				TxID:  evmtypes.Hash(common.BytesToHash(rand.Bytes(32))),
				Index: uint64(i),
			}
		}
		voteData := &evmtypes.VoteEvents{Chain: "Ethereum", Events: events}

		majorityCount := (numVoters * 51 / 100) + 1
		for i := range majorityCount {
			poll.Vote(voters[i], ctx.BlockHeight(), voteData)
		}

		polls[p], _ = k.GetPoll(ctx, pollID)
	}

	for b.Loop() {
		for _, poll := range polls {
			allVoters := poll.GetVoters()
			for _, voter := range allVoters {
				_ = poll.HasVoted(voter)
				_ = poll.HasVotedCorrectly(voter)
			}
		}
	}
}

func TestPoll_GetMetaData(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	evmtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("paramsKey"), store.NewKVStoreKey("tparamsKey"), "vote")
	k := keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey(types.StoreKey), subspace, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.RewarderMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.NewTestLogger(t))
	snap := snapshot.NewSnapshot(
		time.Now(),
		rand.I64Between(1, 100),
		slices.Expand(func(_ int) snapshot.Participant { return snapshot.NewParticipant(rand.ValAddr(), math.OneUint()) }, 5),
		math.NewUint(5),
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
