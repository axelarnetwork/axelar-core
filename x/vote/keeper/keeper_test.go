package keeper

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/utils"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	bcMock "github.com/axelarnetwork/axelar-core/x/broadcast/exported/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
)

var (
	stringGen = testutils.RandStrings(5, 50).Distinct()
)

func init() {
	cdc := testutils.Codec()
	cdc.RegisterConcrete(&mock.MsgVoteMock{}, "mockVote", nil)
	cdc.RegisterConcrete("", "string", nil)
}

type testSetup struct {
	Keeper      Keeper
	Ctx         sdk.Context
	Broadcaster *bcMock.BroadcasterMock
	Snapshotter *snapMock.SnapshotterMock
	// used by the snapshotter when returning a snapshot
	ValidatorSet []snapshot.Validator
	Timeout      context.Context
	cancel       context.CancelFunc
}

func setup() *testSetup {
	setup := &testSetup{Ctx: sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())}
	setup.Snapshotter = &snapMock.SnapshotterMock{
		GetLatestRoundFunc: func(sdk.Context) int64 {
			return testutils.RandIntBetween(1, 10000)
		},
		GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) {
			totalPower := sdk.ZeroInt()
			for _, v := range setup.ValidatorSet {
				totalPower = totalPower.AddRaw(v.GetConsensusPower())
			}
			return snapshot.Snapshot{Validators: setup.ValidatorSet, TotalPower: totalPower}, true
		},
	}
	setup.Broadcaster = &bcMock.BroadcasterMock{BroadcastFunc: func(sdk.Context, []broadcast.MsgWithSenderSetter) error {
		setup.cancel()
		return nil
	}}
	setup.Keeper = NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(stringGen.Next()), store.NewSubjectiveStore(),
		setup.Snapshotter, setup.Broadcaster)
	return setup
}

func (s *testSetup) NewTimeout(t time.Duration) {
	s.Timeout, s.cancel = context.WithTimeout(context.Background(), t)
}

// no error on initializing new poll
func TestKeeper_InitPoll_NoError(t *testing.T) {
	s := setup()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, randPoll()))
}

// error when initializing poll with same id as existing poll
func TestKeeper_InitPoll_SameIdReturnError(t *testing.T) {
	s := setup()

	poll := randPoll()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	assert.Error(t, s.Keeper.InitPoll(s.Ctx, poll))
}

// vote for existing poll is broadcast exactly once
func TestKeeper_Vote_OnNextBroadcast(t *testing.T) {
	s := setup()

	poll := randPoll()
	vote := randVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	assert.NoError(t, s.Keeper.RecordVote(s.Ctx, vote))

	// give go a chance to switch context, because broadcast needs to be done on a different thread
	s.NewTimeout(10 * time.Millisecond)
	s.Keeper.SendVotes(s.Ctx)
	<-s.Timeout.Done()

	assert.Equal(t, 1, len(s.Broadcaster.BroadcastCalls()))
	assert.Equal(t, 1, len(s.Broadcaster.BroadcastCalls()[0].Msgs))
	m := s.Broadcaster.BroadcastCalls()[0].Msgs[0]
	assert.Equal(t, vote, m.(exported.MsgVote))
}

// error when voting for unknown poll, no polls initialized
func TestKeeper_Vote_On_NoPolls_ReturnError(t *testing.T) {
	s := setup()

	poll := randPoll()
	vote := randVoteForPoll(poll)
	assert.Error(t, s.Keeper.RecordVote(s.Ctx, vote))
}

// error when voting where poll id matches none of the existing polls
func TestKeeper_Vote_PollIdMismatch_ReturnError(t *testing.T) {
	s := setup()

	initializedPoll := randPoll()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, initializedPoll))

	notInitializedPoll := randPoll()
	vote := randVoteForPoll(notInitializedPoll)

	assert.Error(t, s.Keeper.RecordVote(s.Ctx, vote))
}

// send two votes on first broadcast and one vote on second broadcast
func TestKeeper_Vote_VotesNotRepeatedInConsecutiveBroadcasts(t *testing.T) {
	s := setup()

	poll1 := randPoll()
	poll2 := randPoll()
	poll3 := randPoll()

	voteForPoll1 := randVoteForPoll(poll1)
	voteForPoll2 := randVoteForPoll(poll2)
	voteForPoll3 := randVoteForPoll(poll3)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll1))
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll2))
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll3))

	assert.NoError(t, s.Keeper.RecordVote(s.Ctx, voteForPoll1))
	assert.NoError(t, s.Keeper.RecordVote(s.Ctx, voteForPoll2))

	// give go a chance to switch context, because broadcast needs to be done on a different thread
	s.NewTimeout(10 * time.Millisecond)
	s.Keeper.SendVotes(s.Ctx)
	<-s.Timeout.Done()

	assert.NoError(t, s.Keeper.RecordVote(s.Ctx, voteForPoll3))

	s.Keeper.SendVotes(s.Ctx)

	// give go a chance to switch context, because broadcast needs to be done on a different thread
	s.NewTimeout(10 * time.Millisecond)
	s.Keeper.SendVotes(s.Ctx)
	<-s.Timeout.Done()

	// assert correct votes
	assert.Equal(t, 2, len(s.Broadcaster.BroadcastCalls()))
	assert.Equal(t, 2, len(s.Broadcaster.BroadcastCalls()[0].Msgs))
	assert.Equal(t, 1, len(s.Broadcaster.BroadcastCalls()[1].Msgs))

	// assert correct votes
	assert.Equal(t, voteForPoll1, s.Broadcaster.BroadcastCalls()[0].Msgs[0])
	assert.Equal(t, voteForPoll2, s.Broadcaster.BroadcastCalls()[0].Msgs[1])
	assert.Equal(t, voteForPoll3, s.Broadcaster.BroadcastCalls()[1].Msgs[0])
}

// error when voting on same poll multiple times
func TestKeeper_Vote_MultipleTimes_ReturnError(t *testing.T) {
	s := setup()

	poll := randPoll()
	vote1 := randVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	assert.NoError(t, s.Keeper.RecordVote(s.Ctx, vote1))

	// submit vote1 again
	assert.Error(t, s.Keeper.RecordVote(s.Ctx, vote1))

	// same poll, different data
	vote2 := vote1
	vote2.DataVal = stringGen.Next()
	assert.Error(t, s.Keeper.RecordVote(s.Ctx, vote2))

	// same poll, different data
	vote3 := vote1
	vote3.DataVal = stringGen.Next()
	assert.Error(t, s.Keeper.RecordVote(s.Ctx, vote3))
}

// send no broadcast when there are no votes
func TestKeeper_Vote_noVotes_NoBroadcast(t *testing.T) {
	s := setup()

	poll1 := randPoll()
	poll2 := randPoll()
	poll3 := randPoll()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll1))
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll2))
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll3))

	// give go a chance to switch context, because broadcast needs to be done on a different thread
	s.NewTimeout(10 * time.Millisecond)
	s.Keeper.SendVotes(s.Ctx)
	<-s.Timeout.Done()

	assert.Equal(t, 0, len(s.Broadcaster.BroadcastCalls()))
}

// error when tallying non-existing poll
func TestKeeper_TallyVote_NonExistingPoll_ReturnError(t *testing.T) {
	s := setup()

	poll := randPoll()
	vote := randVote()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote))
}

// error when tallied vote comes from unauthorized voter
func TestKeeper_TallyVote_UnknownVoter_ReturnError(t *testing.T) {
	s := setup()
	// proxy is unknown
	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return nil }

	poll := randPoll()
	vote := randVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote))
}

// tally vote no winner
func TestKeeper_TallyVote_NoWinner(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	minorityPower := newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(0, 200))
	majorityPower := newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(3/2*minorityPower.GetConsensusPower()+1, 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return minorityPower.GetOperator() }

	poll := randPoll()
	vote := randVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	err := s.Keeper.TallyVote(s.Ctx, vote)
	res := s.Keeper.Result(s.Ctx, poll)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

// tally vote with winner
func TestKeeper_TallyVote_WithWinner(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	minorityPower := newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(0, 200))
	majorityPower := newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(3/2*minorityPower.GetConsensusPower()+1, 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return majorityPower.GetOperator() }
	poll := randPoll()
	vote := randVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	err := s.Keeper.TallyVote(s.Ctx, vote)
	res := s.Keeper.Result(s.Ctx, poll)
	assert.NoError(t, err)
	assert.Equal(t, vote.Data(), res)
}

// error when tallying second vote from same validator
func TestKeeper_TallyVote_TwoVotesFromSameValidator_ReturnError(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(0, 1000))}

	// return same validator for all votes
	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return s.ValidatorSet[0].GetOperator() }

	poll := randPoll()
	vote1 := randVoteForPoll(poll)
	vote2 := randVoteForPoll(poll)
	vote3 := randVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))
	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote1))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote2))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote3))
}

// tally multiple votes until poll is decided
func TestKeeper_TallyVote_MultipleVotesUntilDecision(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{
		// ensure first validator does not have majority voting power
		newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(0, 100)),
		newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(100, 200)),
		newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(100, 200)),
		newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(100, 200)),
		newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(100, 200)),
	}

	poll := randPoll()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))

	vote := randVoteForPoll(poll)
	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return s.ValidatorSet[0].GetOperator() }

	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote))
	assert.Nil(t, s.Keeper.Result(s.Ctx, poll))

	var pollDecided bool
	for i, val := range s.ValidatorSet {
		if i == 0 {
			continue
		}
		s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return val.GetOperator() }
		assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote))
		pollDecided = pollDecided || s.Keeper.Result(s.Ctx, poll) != nil
	}

	assert.Equal(t, vote.Data(), s.Keeper.Result(s.Ctx, poll))
}

// tally vote for already decided vote
func TestKeeper_TallyVote_ForDecidedPoll(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	minorityPower := newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(0, 200))
	majorityPower := newValidator(sdk.ValAddress(stringGen.Next()), testutils.RandIntBetween(3/2*minorityPower.GetConsensusPower()+1, 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	poll := randPoll()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll))

	vote1 := randVoteForPoll(poll)
	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return majorityPower.GetOperator() }

	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote1))
	assert.Equal(t, vote1.Data(), s.Keeper.Result(s.Ctx, poll))

	vote2 := randVoteForPoll(poll)
	assert.NotEqual(t, vote1.Data(), vote2.Data())
	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return minorityPower.GetOperator() }

	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote2))
	assert.Equal(t, vote1.Data(), s.Keeper.Result(s.Ctx, poll))
}

func randVoteForPoll(poll exported.PollMeta) *mock.MsgVoteMock {
	vote := randVote()
	vote.PollVal = poll
	return vote
}

func randVote() *mock.MsgVoteMock {
	return &mock.MsgVoteMock{PollVal: randPoll(), DataVal: stringGen.Next(), Sender: sdk.AccAddress(stringGen.Next())}
}

func randPoll() exported.PollMeta {
	return exported.PollMeta{Module: stringGen.Next(), Type: stringGen.Next(), ID: stringGen.Next()}
}

func newValidator(address sdk.ValAddress, power int64) *snapMock.ValidatorMock {
	return &snapMock.ValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power }}
}
