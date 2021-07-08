package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

var stringGen = rand.Strings(5, 50).Distinct()

type testSetup struct {
	Keeper      keeper.Keeper
	Ctx         sdk.Context
	Snapshotter *snapMock.SnapshotterMock
	// used by the snapshotter when returning a snapshot
	ValidatorSet []snapshot.Validator
	Timeout      context.Context
	cancel       context.CancelFunc
}

func setup() *testSetup {
	encCfg := app.MakeEncodingConfig()
	encCfg.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&gogoprototypes.StringValue{},
	)

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
		GetPrincipalFunc: func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return rand.Bytes(sdk.AddrLen) },
	}

	setup.Keeper = keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey(stringGen.Next()), setup.Snapshotter)
	setup.Keeper.SetDefaultVotingThreshold(setup.Ctx, types.DefaultGenesisState().VotingThreshold)
	return setup
}

func (s *testSetup) NewTimeout(t time.Duration) {
	s.Timeout, s.cancel = context.WithTimeout(context.Background(), t)
}

func TestInitPoll(t *testing.T) {
	t.Run("should create a new poll", testutils.Func(func(t *testing.T) {
		s := setup()

		pollKey := randomPollKey()
		snapshotCounter := int64(100)
		expireAt := int64(0)

		assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, snapshotCounter, expireAt))

		expected := types.NewPoll(pollKey, snapshotCounter, expireAt, types.DefaultGenesisState().VotingThreshold)
		actual := s.Keeper.GetPoll(s.Ctx, pollKey)
		assert.Equal(t, expected, *actual)
	}))

	t.Run("should return error if poll with same key exists and has not expired yet", testutils.Func(func(t *testing.T) {
		s := setup()
		pollKey := randomPollKey()
		snapshotCounter := int64(100)
		expireAt := int64(0)

		assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, snapshotCounter, expireAt))
		assert.Error(t, s.Keeper.InitPoll(s.Ctx, pollKey, snapshotCounter, expireAt))
	}))

	t.Run("should create a new poll if poll with same key exists and has already expired", testutils.Func(func(t *testing.T) {
		s := setup()
		pollKey := randomPollKey()
		snapshotCounter1 := int64(100)
		snapshotCounter2 := int64(101)
		expireAt1 := int64(10)
		expireAt2 := int64(20)

		assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, snapshotCounter1, expireAt1))
		assert.NoError(t, s.Keeper.InitPoll(s.Ctx.WithBlockHeight(expireAt1), pollKey, snapshotCounter2, expireAt2))

		expected := types.NewPoll(pollKey, snapshotCounter2, expireAt2, types.DefaultGenesisState().VotingThreshold)
		actual := s.Keeper.GetPoll(s.Ctx, pollKey)
		assert.Equal(t, expected, *actual)
	}))
}

// error when tallying non-existing poll
func TestTallyVote_NonExistingPoll_ReturnError(t *testing.T) {
	s := setup()

	pollKey := randomPollKey()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, rand.PosI64(), 0))
	_, err := s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, randomData())
	assert.Error(t, err)
}

// error when tallied vote comes from unauthorized voter
func TestTallyVote_UnknownVoter_ReturnError(t *testing.T) {
	s := setup()
	// proxy is unknown
	s.Snapshotter.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress { return nil }

	pollKey := randomPollKey()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, rand.PosI64(), 0))
	_, err := s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, randomData())
	assert.Error(t, err)
}

// tally vote no winner
func TestTallyVote_NoWinner(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetDefaultVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 200))
	majorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Snapshotter.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
		return minorityPower.GetSDKValidator().GetOperator()
	}

	pollKey := randomPollKey()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, rand.PosI64(), 0))
	poll, err := s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, randomData())
	assert.NoError(t, err)
	assert.Nil(t, poll.GetResult())
}

// tally vote with winner
func TestTallyVote_WithWinner(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetDefaultVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 200))
	majorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	s.Snapshotter.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
		return majorityPower.GetSDKValidator().GetOperator()
	}
	pollKey := randomPollKey()
	data := randomData()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, rand.PosI64(), 0))
	poll, err := s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, data)

	assert.NoError(t, err)
	assert.Equal(t, data, poll.GetResult())
}

// error when tallying second vote from same validator
func TestTallyVote_TwoVotesFromSameValidator_ReturnError(t *testing.T) {
	s := setup()
	s.Keeper.SetDefaultVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 1000))}

	// return same validator for all votes
	s.Snapshotter.GetPrincipalFunc = func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
		return s.ValidatorSet[0].GetSDKValidator().GetOperator()
	}

	pollKey := randomPollKey()
	sender := randomSender()

	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, rand.PosI64(), 0))
	_, err := s.Keeper.TallyVote(s.Ctx, sender, pollKey, randomData())
	assert.NoError(t, err)
	_, err = s.Keeper.TallyVote(s.Ctx, sender, pollKey, randomData())
	assert.Error(t, err)
	_, err = s.Keeper.TallyVote(s.Ctx, sender, pollKey, randomData())
	assert.Error(t, err)
}

// tally multiple votes until poll is decided
func TestTallyVote_MultipleVotesUntilDecision(t *testing.T) {
	s := setup()
	s.Keeper.SetDefaultVotingThreshold(s.Ctx, utils.Threshold{Numerator: 2, Denominator: 3})
	s.ValidatorSet = []snapshot.Validator{
		// ensure first validator does not have majority voting power
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 100)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
		newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(100, 200)),
	}

	pollKey := randomPollKey()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, 100, 0))

	sender := randomSender()
	data := randomData()

	s.Snapshotter.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
		return s.ValidatorSet[0].GetSDKValidator().GetOperator()
	}

	poll, err := s.Keeper.TallyVote(s.Ctx, sender, pollKey, data)
	assert.NoError(t, err)
	assert.Nil(t, poll.GetResult())

	var pollDecided bool
	for i, val := range s.ValidatorSet {
		if i == 0 {
			continue
		}
		s.Snapshotter.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
			return val.GetSDKValidator().GetOperator()
		}
		poll, err = s.Keeper.TallyVote(s.Ctx, sender, pollKey, data)
		assert.NoError(t, err)
		pollDecided = pollDecided || poll.GetResult() != nil
	}

	assert.Equal(t, data, s.Keeper.GetPoll(s.Ctx, pollKey).GetResult())
}

// tally vote for already decided vote
func TestTallyVote_ForDecidedPoll(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 2, Denominator: 3}
	s.Keeper.SetDefaultVotingThreshold(s.Ctx, threshold)
	minorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(1, 200))
	majorityPower := newValidator(rand.Bytes(sdk.AddrLen), rand.I64Between(calcMajorityLowerLimit(threshold, minorityPower), 1000))
	s.ValidatorSet = []snapshot.Validator{minorityPower, majorityPower}

	pollKey := randomPollKey()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, 100, 0))

	s.Snapshotter.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
		return majorityPower.GetSDKValidator().GetOperator()
	}

	data1 := randomData()
	poll, err := s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, data1)
	assert.NoError(t, err)
	assert.Equal(t, data1, poll.GetResult())

	s.Snapshotter.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
		return minorityPower.GetSDKValidator().GetOperator()
	}

	data2 := randomData()
	poll, err = s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, data2)
	assert.NoError(t, err)
	// does not change outcome
	assert.Equal(t, data1, poll.GetResult())
}

func TestTallyVote_FailedPoll(t *testing.T) {
	s := setup()
	threshold := utils.Threshold{Numerator: 1, Denominator: 2}
	s.Keeper.SetDefaultVotingThreshold(s.Ctx, threshold)
	validatorPower := rand.I64Between(1, 200)
	validator1 := newValidator(rand.Bytes(sdk.AddrLen), validatorPower)
	validator2 := newValidator(rand.Bytes(sdk.AddrLen), validatorPower)
	s.ValidatorSet = []snapshot.Validator{validator1, validator2}

	pollKey := randomPollKey()
	assert.NoError(t, s.Keeper.InitPoll(s.Ctx, pollKey, 100, 0))

	s.Snapshotter.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return validator1.GetSDKValidator().GetOperator() }
	poll, err := s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, randomData())
	assert.NoError(t, err)
	assert.Nil(t, poll.GetResult())
	assert.False(t, poll.Failed)

	s.Snapshotter.GetPrincipalFunc = func(sdk.Context, sdk.AccAddress) sdk.ValAddress { return validator2.GetSDKValidator().GetOperator() }
	poll, err = s.Keeper.TallyVote(s.Ctx, randomSender(), pollKey, randomData())
	assert.NoError(t, err)
	assert.Nil(t, poll.GetResult())
	assert.True(t, poll.Failed)
}

func randomData() codec.ProtoMarshaler {
	return &gogoprototypes.StringValue{Value: stringGen.Next()}
}

func randomSender() sdk.AccAddress {
	return rand.Bytes(sdk.AddrLen)
}

func randomPollKey() exported.PollKey {
	return exported.NewPollKey(stringGen.Next(), stringGen.Next())
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
		majorityLowerLimit++
	}
	return majorityLowerLimit
}
