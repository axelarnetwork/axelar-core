package keeper

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	bcMock "github.com/axelarnetwork/axelar-core/x/broadcast/exported/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
)

var stringGen = rand.Strings(5, 50).Distinct()

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
		GetLatestCounterFunc: func(sdk.Context) int64 { return rand.I64Between(1, 10000) },
		GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) {
			totalPower := sdk.ZeroInt()
			for _, v := range setup.ValidatorSet {
				totalPower = totalPower.AddRaw(v.GetConsensusPower())
			}
			return snapshot.Snapshot{Validators: setup.ValidatorSet, TotalPower: totalPower}, true
		},
	}
	setup.Broadcaster = &bcMock.BroadcasterMock{}
	setup.Keeper = NewKeeper(testutils.Codec(), sdk.NewKVStoreKey(stringGen.Next()), setup.Snapshotter, setup.Broadcaster)
	return setup
}

func (s *testSetup) NewTimeout(t time.Duration) {
	s.Timeout, s.cancel = context.WithTimeout(context.Background(), t)
}

// no error on initializing new poll
func TestKeeper_InitPoll_NoError(t *testing.T) {
	s := setup()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, randomPoll(), 100))
}

// error when initializing poll with same id as existing poll
func TestKeeper_InitPoll_SameIdReturnError(t *testing.T) {
	s := setup()

	poll := randomPoll()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))
	assert.Error(t, s.Keeper.InitPoll(s.Ctx, poll, 100))
}

// error when tallying non-existing poll
func TestKeeper_TallyVote_NonExistingPoll_ReturnError(t *testing.T) {
	s := setup()

	poll := randomPoll()
	vote := randomVote()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote.Sender, vote.Poll(), vote.Data()))
}

// error when tallied vote comes from unauthorized voter
func TestKeeper_TallyVote_UnknownVoter_ReturnError(t *testing.T) {
	s := setup()
	// proxy is unknown
	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return nil }

	poll := randomPoll()
	vote := randomVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote.Sender, vote.Poll(), vote.Data()))
}

// tally vote no winner
func TestKeeper_TallyVote_NoWinner(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(1, 200))
	majorityPower := newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return minorityPower.GetOperator() }

	poll := randomPoll()
	vote := randomVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))
	err := s.Keeper.TallyVote(s.Ctx, vote.Sender, vote.Poll(), vote.Data())
	res := s.Keeper.Result(s.Ctx, poll)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

// tally vote with winner
func TestKeeper_TallyVote_WithWinner(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(1, 200))
	majorityPower := newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return majorityPower.GetOperator() }
	poll := randomPoll()
	vote := randomVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))
	err := s.Keeper.TallyVote(s.Ctx, vote.Sender, vote.Poll(), vote.Data())
	res := s.Keeper.Result(s.Ctx, poll)
	assert.NoError(t, err)
	assert.Equal(t, vote.Data(), res)
}

// error when tallying second vote from same validator
func TestKeeper_TallyVote_TwoVotesFromSameValidator_ReturnError(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(1, 1000))}

	// return same validator for all votes
	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return s.ValidatorSet[0].GetOperator() }

	poll := randomPoll()
	vote1 := randomVoteForPoll(poll)
	vote2 := randomVoteForPoll(poll)
	vote3 := randomVoteForPoll(poll)

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))
	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote1.Sender, vote1.Poll(), vote1.Data()))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote2.Sender, vote2.Poll(), vote2.Data()))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, vote3.Sender, vote3.Poll(), vote3.Data()))
}

// tally multiple votes until poll is decided
func TestKeeper_TallyVote_MultipleVotesUntilDecision(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{
		// ensure first validator does not have majority voting power
		newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(1, 100)),
		newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(100, 200)),
		newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(100, 200)),
		newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(100, 200)),
		newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(100, 200)),
	}

	poll := randomPoll()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))

	vote := randomVoteForPoll(poll)
	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return s.ValidatorSet[0].GetOperator() }

	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote.Sender, vote.Poll(), vote.Data()))
	assert.Nil(t, s.Keeper.Result(s.Ctx, poll))

	var pollDecided bool
	for i, val := range s.ValidatorSet {
		if i == 0 {
			continue
		}
		s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return val.GetOperator() }
		assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote.Sender, vote.Poll(), vote.Data()))
		pollDecided = pollDecided || s.Keeper.Result(s.Ctx, poll) != nil
	}

	assert.Equal(t, vote.Data(), s.Keeper.Result(s.Ctx, poll))
}

// tally vote for already decided vote
func TestKeeper_TallyVote_ForDecidedPoll(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(1, 200))
	majorityPower := newValidator(sdk.ValAddress(stringGen.Next()), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	poll := randomPoll()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))

	vote1 := randomVoteForPoll(poll)
	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return majorityPower.GetOperator() }

	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote1.Sender, vote1.Poll(), vote1.Data()))
	assert.Equal(t, vote1.Data(), s.Keeper.Result(s.Ctx, poll))

	vote2 := randomVoteForPoll(poll)
	assert.NotEqual(t, vote1.Data(), vote2.Data())
	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return minorityPower.GetOperator() }

	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, vote2.Sender, vote2.Poll(), vote2.Data()))
	assert.Equal(t, vote1.Data(), s.Keeper.Result(s.Ctx, poll))
}

func randomVoteForPoll(poll exported.PollMeta) *mock.MsgVoteMock {
	vote := randomVote()
	vote.PollVal = poll
	return vote
}

func randomVote() *mock.MsgVoteMock {
	return &mock.MsgVoteMock{PollVal: randomPoll(), DataVal: stringGen.Next(), Sender: sdk.AccAddress(stringGen.Next())}
}

func randomPoll() exported.PollMeta {
	return exported.NewPollMeta(stringGen.Next(), stringGen.Next(), stringGen.Next())
}

func newValidator(address sdk.ValAddress, power int64) *snapMock.ValidatorMock {
	return &snapMock.ValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power }}
}

func calcMajorityLowerLimit(threshold utils.Threshold, minorityPower *snapMock.ValidatorMock) int64 {
	minorityShare := threshold.Denominator - threshold.Numerator
	majorityShare := threshold.Numerator
	majorityLowerLimit := minorityPower.GetConsensusPower() / minorityShare * majorityShare
	// Due to integer division the lower limit might be underestimated by up to 2
	for threshold.IsMet(sdk.NewInt(majorityLowerLimit), sdk.NewInt(majorityLowerLimit+minorityPower.GetConsensusPower())) {
		majorityLowerLimit += 1
	}
	return majorityLowerLimit
}
