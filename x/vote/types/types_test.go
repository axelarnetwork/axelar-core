package types_test

import (
	"fmt"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
)

func TestNewTalliedVote(t *testing.T) {
	t.Run("panic on nil data", func(t *testing.T) {
		assert.Panics(t, func() {
			types.NewTalliedVote(rand.ValAddr(), rand.PosI64(), nil)
		})
	})

	t.Run("panic on nil voter", func(t *testing.T) {
		assert.Panics(t, func() {
			types.NewTalliedVote(nil, rand.PosI64(), &gogoprototypes.BoolValue{Value: true})
		})
	})
}

func TestTalliedVote_Marshaling(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	cdc := encCfg.Marshaler

	output := tofnd.KeygenOutput{PubKey: []byte("a public key"), GroupRecoverInfo: []byte{0}, PrivateRecoverInfo: []byte{0, 1, 2, 3}}
	data := tofnd.MessageOut_KeygenResult{KeygenResultData: &tofnd.MessageOut_KeygenResult_Data{Data: &output}}
	vote := types.NewTalliedVote(rand.ValAddr(), 23, &data)

	bz := cdc.MustMarshalLengthPrefixed(&vote)
	var actual types.TalliedVote
	cdc.MustUnmarshalLengthPrefixed(bz, &actual)

	assert.Equal(t, vote, actual)

	bz = cdc.MustMarshalJSON(&vote)
	var actual2 types.TalliedVote
	cdc.MustUnmarshalJSON(bz, &actual2)

	assert.Equal(t, vote.Tally, actual2.Tally)
	assert.Equal(t, vote.Data.GetCachedValue(), actual2.Data.GetCachedValue())
}

func TestPoll_Expiry(t *testing.T) {
	setup := func() exported.PollMetadata {
		key := exported.NewPollKey(rand.StrBetween(5, 20), rand.StrBetween(5, 20))
		return types.NewPollMetaData(key, types.DefaultGenesisState().VotingThreshold, []exported.Voter{}, sdk.ZeroInt())
	}
	repeats := 20
	notExpiredStates := []exported.PollState{exported.NonExistent, exported.Pending, exported.Completed, exported.Failed}

	t.Run("poll is not expired", testutils.Func(func(t *testing.T) {
		metadata := setup()
		initialState := notExpiredStates[rand.I64Between(0, int64(len(notExpiredStates)))]
		metadata.State = initialState
		expiry := rand.PosI64()
		metadata.ExpiresAt = expiry
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(expiry - rand.I64Between(0, expiry))

		poll := types.NewPoll(ctx, metadata, &mock.StoreMock{}, &mock.RewarderMock{})

		assert.True(t, poll.Is(initialState))
	}).Repeat(repeats))

	t.Run("pending poll expires", testutils.Func(func(t *testing.T) {
		metadata := setup()
		metadata.State = exported.Pending
		expiry := rand.I64Between(0, 1000000)
		metadata.ExpiresAt = expiry
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(expiry + rand.I64Between(1, 1000000))

		poll := types.NewPoll(ctx, metadata, &mock.StoreMock{}, &mock.RewarderMock{})

		assert.True(t, poll.Is(exported.Pending))
		assert.True(t, poll.Is(exported.Expired))
	}).Repeat(repeats))

	t.Run("not-pending poll does not expire", testutils.Func(func(t *testing.T) {
		initialState := notExpiredStates[rand.I64GenBetween(0, int64(len(notExpiredStates))).
			Where(func(i int64) bool { return i != int64(exported.Pending) }).
			Next()]
		metadata := setup()
		metadata.State = initialState
		expiry := rand.I64Between(0, 1000000)
		metadata.ExpiresAt = expiry
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(expiry + rand.I64Between(1, 1000000))
		poll := types.NewPoll(ctx, metadata, &mock.StoreMock{}, &mock.RewarderMock{})

		assert.True(t, poll.Is(initialState))
		assert.False(t, poll.Is(exported.Expired))
	}).Repeat(repeats))
}

func TestPoll_Is(t *testing.T) {
	for _, state := range exported.PollState_value {
		poll := types.Poll{PollMetadata: exported.PollMetadata{State: exported.PollState(state)}}

		assert.True(t, poll.Is(exported.PollState(state)))

		for _, otherState := range exported.PollState_value {
			if otherState == state {
				continue
			}
			assert.False(t, poll.Is(exported.PollState(otherState)), "poll: %s, other: %s", poll.State, exported.PollState(otherState))
		}
	}
}

func TestPoll_Vote(t *testing.T) {
	var (
		votingPowers     map[string]int64
		totalVotingPower sdk.Int
	)

	setup := func(metadata exported.PollMetadata, currBlockHeight int64) *types.Poll {
		votingPowers = randomEvenVotingPowers()

		totalVotingPower = sdk.ZeroInt()
		for _, votingPower := range votingPowers {
			totalVotingPower = totalVotingPower.AddRaw(votingPower)
		}

		allVotes := make(map[string]types.TalliedVote)
		hasVoted := make(map[string]bool)
		store := &mock.StoreMock{
			SetVoteFunc: func(addr sdk.ValAddress, v types.TalliedVote) {
				hasVoted[addr.String()] = true
				allVotes[v.Hash()] = v
			},
			GetVoteFunc: func(h string) (types.TalliedVote, bool) {
				vote, ok := allVotes[h]
				return vote, ok
			},
			GetVotesFunc:    func() []types.TalliedVote { return getValues(allVotes) },
			HasVotedFunc:    func(addr sdk.ValAddress) bool { return hasVoted[addr.String()] },
			SetMetadataFunc: func(exported.PollMetadata) {},
		}

		for voterAddress, votingPower := range votingPowers {
			validator, err := sdk.ValAddressFromBech32(voterAddress)
			if err != nil {
				panic(err)
			}

			metadata.Voters = append(metadata.Voters, exported.Voter{Validator: validator, VotingPower: votingPower})
		}
		metadata.TotalVotingPower = totalVotingPower
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(currBlockHeight)

		return types.NewPoll(ctx, metadata, store, &mock.RewarderMock{}).WithLogger(log.TestingLogger())
	}
	repeats := 20

	t.Run("poll nonexistent", testutils.Func(func(t *testing.T) {
		poll := setup(exported.PollMetadata{State: exported.NonExistent}, rand.PosI64())

		voter := rand.ValAddr()
		votingPowers[voter.String()] = rand.PosI64()
		totalVotingPower = sdk.NewInt(votingPowers[voter.String()]).MulRaw(10)

		assert.Error(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("poll already completed", testutils.Func(func(t *testing.T) {
		result, _ := codectypes.NewAnyWithValue(&gogoprototypes.StringValue{Value: rand.Str(10)})
		poll := setup(exported.PollMetadata{State: exported.Completed, Result: result}, rand.PosI64())

		voter := rand.ValAddr()
		votingPowers[voter.String()] = rand.PosI64()
		totalVotingPower = sdk.NewInt(votingPowers[voter.String()]).MulRaw(10)

		assert.NoError(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("poll already failed", testutils.Func(func(t *testing.T) {
		poll := setup(exported.PollMetadata{State: exported.Failed}, rand.PosI64())

		voter := rand.ValAddr()
		votingPowers[voter.String()] = rand.PosI64()
		totalVotingPower = sdk.NewInt(votingPowers[voter.String()]).MulRaw(10)

		assert.NoError(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("voter unknown", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())

		voterAddr := rand.ValAddr()
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.Error(t, poll.Vote(voterAddr, voteValue))
	}).Repeat(repeats))

	t.Run("correct vote no completion", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		metadata.VotingThreshold = utils.OneThreshold
		poll := setup(metadata, rand.PosI64())

		voterAddr := poll.Voters[rand.I64Between(0, int64(len(poll.Voters)))].Validator
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.NoError(t, poll.Vote(voterAddr, voteValue))
		assert.True(t, poll.Is(exported.Pending))
	}).Repeat(repeats))

	t.Run("vote after expiry", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		currBlockHeight := metadata.ExpiresAt + rand.I64Between(1, 1000000)
		poll := setup(metadata, currBlockHeight)

		assert.True(t, poll.Is(exported.Expired))
		assert.True(t, poll.Is(exported.Pending))

		voterAddr := poll.Voters[rand.I64Between(0, int64(len(poll.Voters)))].Validator
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.NoError(t, poll.Vote(voterAddr, voteValue))
	}).Repeat(repeats))

	t.Run("already voted", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())

		voterAddr := poll.Voters[rand.I64Between(0, int64(len(poll.Voters)))].Validator
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.NoError(t, poll.Vote(voterAddr, voteValue))
		assert.Error(t, poll.Vote(voterAddr, voteValue))
	}).Repeat(repeats))

	t.Run("multiple votes until completion", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())

		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
		for voter := range votingPowers {
			addr, _ := sdk.ValAddressFromBech32(voter)
			assert.NoError(t, poll.Vote(addr, voteValue))
		}

		assert.True(t, poll.Is(exported.Completed))
		assert.Equal(t, voteValue, poll.GetResult())
	}).Repeat(repeats))

	t.Run("poll fails", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())

		t.Log(len(votingPowers), "votes")
		for voter := range votingPowers {
			addr, _ := sdk.ValAddressFromBech32(voter)
			voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
			assert.NoError(t, poll.Vote(addr, voteValue))
		}

		assert.True(t, poll.Is(exported.Failed), poll.State)
	}).Repeat(repeats))

	t.Run("should complete the poll when total voter count is less than minimum voter count and minimum voter count is not met", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())
		poll.MinVoterCount = int64(len(votingPowers) + 1)

		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
		for voter := range votingPowers {
			addr, _ := sdk.ValAddressFromBech32(voter)
			assert.NoError(t, poll.Vote(addr, voteValue))
		}

		assert.True(t, poll.Is(exported.Completed))
		assert.Equal(t, voteValue, poll.GetResult())
	}).Repeat(repeats))

	t.Run("should not complete the poll when total voter count is greater than or equal to minimum voter count and minimum voter count is not met", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())
		poll.MinVoterCount = int64(len(votingPowers))
		poll.VotingThreshold = utils.ZeroThreshold

		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
		voterCount := int64(0)
		for voter := range votingPowers {
			if voterCount > metadata.MinVoterCount-2 {
				break
			}

			addr, _ := sdk.ValAddressFromBech32(voter)
			assert.NoError(t, poll.Vote(addr, voteValue))
			voterCount++
		}

		assert.True(t, poll.Is(exported.Pending))
	}).Repeat(repeats))
}

func TestPoll_Initialize(t *testing.T) {
	var (
		previousPoll exported.Poll
	)

	store := &mock.StoreMock{
		GetPollFunc:     func(exported.PollKey) exported.Poll { return previousPoll },
		SetMetadataFunc: func(exported.PollMetadata) {},
		DeletePollFunc:  func() {},
	}

	repeats := 20

	testCases := []struct {
		label        string
		previousPoll exported.Poll
		expectError  bool
	}{
		{"poll can be overridden", &voteMock.PollMock{DeleteFunc: func() error { return nil }}, false},
		{"poll can not be overridden", &voteMock.PollMock{DeleteFunc: func() error { return fmt.Errorf("no delete") }}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, testutils.Func(func(t *testing.T) {
			previousPoll = testCase.previousPoll
			ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(rand.PosI64())
			poll := types.NewPoll(ctx, newRandomPollMetadata(), store, &mock.RewarderMock{}).WithLogger(log.TestingLogger())

			if testCase.expectError {
				assert.Error(t, poll.Initialize())
			} else {
				assert.NoError(t, poll.Initialize())
			}
		}).Repeat(repeats))
	}
}

func TestPoll_Delete(t *testing.T) {
	var store *mock.StoreMock
	setup := func(pollState exported.PollState) exported.Poll {
		store = &mock.StoreMock{DeletePollFunc: func() {}}
		metadata := newRandomPollMetadata()
		metadata.State = pollState
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger()).WithBlockHeight(rand.I64Between(0, metadata.ExpiresAt))

		return types.NewPoll(ctx, metadata, store, &mock.RewarderMock{}).WithLogger(log.TestingLogger())
	}

	t.Run("nonexistent", func(t *testing.T) {
		poll := setup(exported.NonExistent)
		assert.NoError(t, poll.Delete())
		assert.Len(t, store.DeletePollCalls(), 0)
	})

	testCases := []struct {
		label            string
		pollState        exported.PollState
		deleteSuccessful bool
	}{
		{"pending", exported.Pending, false},
		{"completed", exported.Completed, false},
		{"failed", exported.Failed, false},
		{"expired", exported.Expired, false},
		{"allow override", exported.AllowOverride, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			poll := setup(testCase.pollState)

			if testCase.deleteSuccessful {
				assert.NoError(t, poll.Delete())
				assert.Len(t, store.DeletePollCalls(), 1)
			} else {
				assert.Error(t, poll.Delete())
				assert.Len(t, store.DeletePollCalls(), 0)
			}
		})
	}
}

func randomEvenVotingPowers() map[string]int64 {
	votingPowers := make(map[string]int64)

	total := sdk.ZeroInt()
	for i := 0; i < int(rand.I64Between(1, 20)); i++ {
		addr := rand.ValAddr()
		votingPower := rand.I64Between(1, 100)
		votingPowers[addr.String()] = votingPower
		total = total.AddRaw(votingPower)
	}

	// redraw voting power if any one votingPower is greater than half of the total
	for _, votingPower := range votingPowers {
		if total.QuoRaw(2).LT(sdk.NewInt(votingPower)) {
			return randomEvenVotingPowers()
		}
	}

	return votingPowers
}

func getValues(m map[string]types.TalliedVote) []types.TalliedVote {
	votes := make([]types.TalliedVote, 0, len(m))
	for _, vote := range m {
		votes = append(votes, vote)
	}
	return votes
}

func newRandomPollMetadata() exported.PollMetadata {
	key := exported.NewPollKey(rand.StrBetween(5, 20), rand.StrBetween(5, 20))
	poll := types.NewPollMetaData(key, types.DefaultGenesisState().VotingThreshold, []exported.Voter{}, sdk.ZeroInt())
	poll.ExpiresAt = rand.I64Between(1, 1000000)
	return poll
}
