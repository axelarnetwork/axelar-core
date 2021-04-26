package keeper

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	bcMock "github.com/axelarnetwork/axelar-core/x/broadcast/exported/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var stringGen = rand.Strings(5, 50).Distinct()

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
	cdc := testutils.MakeEncodingConfig().Amino
	cdc.RegisterConcrete("", "string", nil)

	setup := &testSetup{Ctx: sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())}
	setup.Snapshotter = &snapMock.SnapshotterMock{
		GetLatestCounterFunc: func(sdk.Context) int64 { return rand.I64Between(1, 10000) },
		GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) {
			totalShareCount := sdk.ZeroInt()
			for _, v := range setup.ValidatorSet {
				totalShareCount = totalShareCount.AddRaw(v.ShareCount)
			}
			return snapshot.Snapshot{Validators: setup.ValidatorSet, TotalShareCount: totalShareCount}, true
		},
	}
	setup.Broadcaster = &bcMock.BroadcasterMock{
		GetPrincipalFunc: func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return rand.Bytes(sdk.AddrLen) },
	}
	setup.Keeper = NewKeeper(cdc, sdk.NewKVStoreKey(stringGen.Next()), setup.Snapshotter, setup.Broadcaster)
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

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, rand.PosI64()))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, randomSender(), poll, randomData()))
}

// error when tallied vote comes from unauthorized voter
func TestKeeper_TallyVote_UnknownVoter_ReturnError(t *testing.T) {
	s := setup()
	// proxy is unknown
	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return nil }

	poll := randomPoll()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, rand.PosI64()))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, randomSender(), poll, randomData()))
}

// tally vote no winner
func TestKeeper_TallyVote_NoWinner(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 200))
	majorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return minorityPower.GetOperator() }

	poll := randomPoll()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, rand.PosI64()))
	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, randomSender(), poll, randomData()))
	assert.Nil(t, s.Keeper.Result(s.Ctx, poll))
}

// tally vote with winner
func TestKeeper_TallyVote_WithWinner(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 200))
	majorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return majorityPower.GetOperator() }
	poll := randomPoll()
	data := randomData()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, rand.PosI64()))
	err := s.Keeper.TallyVote(s.Ctx, randomSender(), poll, data)
	res := s.Keeper.Result(s.Ctx, poll)
	assert.NoError(t, err)
	assert.Equal(t, data, res)
}

// error when tallying second vote from same validator
func TestKeeper_TallyVote_TwoVotesFromSameValidator_ReturnError(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 1000))}

	// return same validator for all votes
	s.Broadcaster.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return s.ValidatorSet[0].GetOperator() }

	poll := randomPoll()
	sender := randomSender()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, rand.PosI64()))
	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, sender, poll, randomData()))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, sender, poll, randomData()))
	assert.Error(t, s.Keeper.TallyVote(s.Ctx, sender, poll, randomData()))
}

// tally multiple votes until poll is decided
func TestKeeper_TallyVote_MultipleVotesUntilDecision(t *testing.T) {
	s := setup()
	s.Keeper.SetVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{
		// ensure first validator does not have majority voting power
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 100)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
	}

	poll := randomPoll()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))

	sender := randomSender()
	data := randomData()

	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return s.ValidatorSet[0].GetOperator() }

	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, sender, poll, data))
	assert.Nil(t, s.Keeper.Result(s.Ctx, poll))

	var pollDecided bool
	for i, val := range s.ValidatorSet {
		if i == 0 {
			continue
		}
		s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return val.GetOperator() }
		assert.NoError(t, s.Keeper.TallyVote(s.Ctx, sender, poll, data))
		pollDecided = pollDecided || s.Keeper.Result(s.Ctx, poll) != nil
	}

	assert.Equal(t, data, s.Keeper.Result(s.Ctx, poll))
}

// tally vote for already decided vote
func TestKeeper_TallyVote_ForDecidedPoll(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 200))
	majorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	poll := randomPoll()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, poll, 100))

	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return majorityPower.GetOperator() }

	data1 := randomData()
	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, randomSender(), poll, data1))
	assert.Equal(t, data1, s.Keeper.Result(s.Ctx, poll))

	s.Broadcaster.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return minorityPower.GetOperator() }

	data2 := randomData()
	assert.NoError(t, s.Keeper.TallyVote(s.Ctx, randomSender(), poll, data2))
	// does not change outcome
	assert.Equal(t, data1, s.Keeper.Result(s.Ctx, poll))
}

func randomData() string {
	return stringGen.Next()
}

func randomSender() sdk.AccAddress {
	return rand.Bytes(sdk.AddrLen)
}

func randomPoll() exported.PollMeta {
	return exported.NewPollMeta(stringGen.Next(), stringGen.Next())
}

func newValidator(address sdk.ValAddress, power int64) snapshot.Validator {
	sdkValidator := &snapMock.SDKValidatorMock{
		GetOperatorFunc: func() sdk.ValAddress { return address },
	}

	return snapshot.NewValidator(sdkValidator, power)
}

func calcMajorityLowerLimit(threshold utils.Threshold, minorityPower snapshot.Validator) int64 {
	minorityShare := threshold.Denominator - threshold.Numerator
	majorityShare := threshold.Numerator
	majorityLowerLimit := minorityPower.ShareCount / minorityShare * majorityShare
	// Due to integer division the lower limit might be underestimated by up to 2
	for threshold.IsMet(sdk.NewInt(majorityLowerLimit), sdk.NewInt(majorityLowerLimit).AddRaw(minorityPower.ShareCount)) {
		majorityLowerLimit += 1
	}
	return majorityLowerLimit
}
