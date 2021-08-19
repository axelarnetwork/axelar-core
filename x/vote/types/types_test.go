package types_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
)

func TestNewTalliedVote(t *testing.T) {
	t.Run("panic on nil data", func(t *testing.T) {
		assert.Panics(t, func() {
			types.NewTalliedVote(rand.Bytes(sdk.AddrLen), rand.PosI64(), nil)
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

	output := tofnd.MessageOut_KeygenResult_KeygenOutput{PubKey: []byte("a public key"), ShareRecoveryInfos: [][]byte{{0, 1, 2, 3}}}
	data := tofnd.MessageOut_KeygenResult{KeygenResultData: &tofnd.MessageOut_KeygenResult_Data{Data: &output}}
	vote := types.NewTalliedVote(rand.Bytes(sdk.AddrLen), 23, &data)

	bz := cdc.MustMarshalBinaryLengthPrefixed(&vote)
	var actual types.TalliedVote
	cdc.MustUnmarshalBinaryLengthPrefixed(bz, &actual)

	assert.Equal(t, vote, actual)

	bz = cdc.MustMarshalJSON(&vote)
	var actual2 types.TalliedVote
	cdc.MustUnmarshalJSON(bz, &actual2)

	assert.Equal(t, vote.Tally, actual2.Tally)
	assert.Equal(t, vote.Data.GetCachedValue(), actual2.Data.GetCachedValue())
}

func TestPoll_CheckExpiry(t *testing.T) {
	setup := func() exported.PollMetadata {
		key := exported.NewPollKey(rand.StrBetween(5, 20), rand.StrBetween(5, 20))
		return types.NewPollMetaData(key, types.DefaultGenesisState().VotingThreshold, rand.PosI64())
	}
	repeats := 20
	notExpiredStates := []exported.PollState{exported.NonExistent, exported.Pending, exported.Completed, exported.Failed}

	t.Run("poll is not expired", testutils.Func(func(t *testing.T) {
		metadata := setup()
		initialState := notExpiredStates[rand.I64Between(0, int64(len(notExpiredStates)))]
		metadata.State = initialState
		expiry := rand.PosI64()
		metadata.ExpiresAt = expiry
		poll := types.NewPoll(metadata, &mock.StoreMock{})

		currBlockHeight := expiry - rand.I64Between(0, expiry)
		poll.CheckExpiry(currBlockHeight)

		assert.True(t, poll.Is(initialState))
	}).Repeat(repeats))

	t.Run("pending poll expires", testutils.Func(func(t *testing.T) {
		metadata := setup()
		metadata.State = exported.Pending
		expiry := rand.I64Between(0, 1000000)
		metadata.ExpiresAt = expiry
		poll := types.NewPoll(metadata, &mock.StoreMock{})

		currBlockHeight := expiry + rand.I64Between(1, 1000000)
		poll.CheckExpiry(currBlockHeight)

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
		poll := types.NewPoll(metadata, &mock.StoreMock{})

		currBlockHeight := expiry + rand.I64Between(1, 1000000)
		poll.CheckExpiry(currBlockHeight)

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
		shareCounts     map[string]int64
		totalShareCount sdk.Int
	)

	setup := func(metadata exported.PollMetadata) *types.PollWithLogging {
		shareCounts = randomEvenShareCounts()
		totalShareCount = sdk.ZeroInt()
		for _, share := range shareCounts {
			totalShareCount = totalShareCount.AddRaw(share)
		}

		allVotes := make(map[string]types.TalliedVote)
		hasVoted := make(map[string]bool)
		store := &mock.StoreMock{
			GetTotalVoterCountFunc: func() int64 { return int64(len(shareCounts)) },
			SetVoteFunc: func(addr sdk.ValAddress, v types.TalliedVote) {
				hasVoted[addr.String()] = true
				allVotes[v.Hash()] = v
			},
			GetVoteFunc: func(h string) (types.TalliedVote, bool) {
				vote, ok := allVotes[h]
				return vote, ok
			},
			GetVotesFunc: func() []types.TalliedVote { return getValues(allVotes) },
			HasVotedFunc: func(addr sdk.ValAddress) bool { return hasVoted[addr.String()] },
			GetShareCountFunc: func(address sdk.ValAddress) (int64, bool) {
				shareCount, ok := shareCounts[address.String()]
				return shareCount, ok
			},
			GetTotalShareCountFunc: func() sdk.Int { return totalShareCount },
			SetMetadataFunc:        func(exported.PollMetadata) {},
		}

		return types.NewPoll(metadata, store).WithLogging(log.TestingLogger())
	}
	repeats := 20

	t.Run("poll nonexistent", testutils.Func(func(t *testing.T) {
		poll := setup(exported.PollMetadata{State: exported.NonExistent})

		voter := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voter.String()] = rand.PosI64()
		totalShareCount = sdk.NewInt(shareCounts[voter.String()]).MulRaw(10)

		assert.Error(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("poll already completed", testutils.Func(func(t *testing.T) {
		result, _ := codectypes.NewAnyWithValue(&gogoprototypes.StringValue{Value: rand.Str(10)})
		poll := setup(exported.PollMetadata{State: exported.Completed, Result: result})

		voter := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voter.String()] = rand.PosI64()
		totalShareCount = sdk.NewInt(shareCounts[voter.String()]).MulRaw(10)

		assert.NoError(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("poll already failed", testutils.Func(func(t *testing.T) {
		poll := setup(exported.PollMetadata{State: exported.Failed})

		voter := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voter.String()] = rand.PosI64()
		totalShareCount = sdk.NewInt(shareCounts[voter.String()]).MulRaw(10)

		assert.NoError(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("voter unknown", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata)

		voterAddr := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.Error(t, poll.Vote(voterAddr, voteValue))
	}).Repeat(repeats))

	t.Run("correct vote no completion", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata)

		voterShareCount := totalShareCount.QuoRaw(int64(len(shareCounts))).Int64() // shareCounts are int64, so this can never be out of bounds
		totalShareCount = totalShareCount.AddRaw(voterShareCount)

		voterAddr := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voterAddr.String()] = voterShareCount
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.NoError(t, poll.Vote(voterAddr, voteValue))
		assert.True(t, poll.Is(exported.Pending))
	}).Repeat(repeats))

	t.Run("vote after expiry", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata)
		currBlockHeight := poll.PollMetadata.ExpiresAt + rand.I64Between(1, 1000000)

		poll.CheckExpiry(currBlockHeight)

		assert.True(t, poll.Is(exported.Expired))
		assert.True(t, poll.Is(exported.Pending))

		voterShareCount := rand.PosI64()
		totalShareCount = totalShareCount.AddRaw(voterShareCount)

		voterAddr := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voterAddr.String()] = voterShareCount
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.NoError(t, poll.Vote(voterAddr, voteValue))
	}).Repeat(repeats))

	t.Run("already voted", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata)

		voterShareCount := totalShareCount.QuoRaw(int64(len(shareCounts))).Int64() // shareCounts are int64, so this can never be out of bounds
		totalShareCount = totalShareCount.AddRaw(voterShareCount)

		voterAddr := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voterAddr.String()] = voterShareCount
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.NoError(t, poll.Vote(voterAddr, voteValue))
		assert.Error(t, poll.Vote(voterAddr, voteValue))
	}).Repeat(repeats))

	t.Run("multiple votes until completion", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata)

		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
		for voter := range shareCounts {
			addr, _ := sdk.ValAddressFromBech32(voter)
			assert.NoError(t, poll.Vote(addr, voteValue))
		}

		assert.True(t, poll.Is(exported.Completed))
		assert.Equal(t, voteValue, poll.GetResult())
	}).Repeat(repeats))

	t.Run("poll fails", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata)

		t.Log(len(shareCounts), "votes")
		for voter := range shareCounts {
			addr, _ := sdk.ValAddressFromBech32(voter)
			voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
			assert.NoError(t, poll.Vote(addr, voteValue))
		}

		assert.True(t, poll.Is(exported.Failed), poll.State)
	}).Repeat(repeats))
}

func TestPoll_Initialize(t *testing.T) {
	var (
		previousPollState exported.PollState
	)

	previousPoll := &voteMock.PollMock{
		IsFunc:     func(state exported.PollState) bool { return state == previousPollState },
		DeleteFunc: func() error { return nil },
	}
	store := &mock.StoreMock{
		GetPollFunc:     func(exported.PollKey) exported.Poll { return previousPoll },
		SetMetadataFunc: func(exported.PollMetadata) {},
		DeletePollFunc:  func() {},
	}

	repeats := 20

	testCases := []struct {
		label             string
		previousPollState exported.PollState
		expectError       bool
	}{
		{"no previous poll exists", exported.NonExistent, false},
		{"pending poll exists", exported.Pending, true},
		{"expired poll exists", exported.Expired, false},
		{"failed poll exists", exported.Failed, false},
		{"completed poll exists", exported.Completed, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, testutils.Func(func(t *testing.T) {
			previousPollState = testCase.previousPollState
			p := types.NewPoll(newRandomPollMetadata(), store)
			poll := p.WithLogging(log.TestingLogger())

			if testCase.expectError {
				assert.Error(t, poll.Initialize())
			} else {
				assert.NoError(t, poll.Initialize())
			}
		}).Repeat(repeats))
	}
}

func TestPollWithLogging_Delete(t *testing.T) {
	var store *mock.StoreMock
	setup := func(pollState exported.PollState) exported.Poll {
		store = &mock.StoreMock{DeletePollFunc: func() {}}
		metadata := newRandomPollMetadata()
		metadata.State = pollState
		return types.NewPoll(metadata, store).WithLogging(log.TestingLogger())
	}

	testCases := []struct {
		label        string
		pollState    exported.PollState
		deleteCalled bool
	}{
		{"nonexistent", exported.NonExistent, false},
		{"pending", exported.Pending, false},
		{"completed", exported.Completed, false},
		{"failed", exported.Failed, true},
		{"expired", exported.Expired, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, testutils.Func(func(t *testing.T) {
			poll := setup(testCase.pollState)

			poll.Delete()

			if testCase.deleteCalled {
				assert.Len(t, store.DeletePollCalls(), 1)
			} else {
				assert.Len(t, store.DeletePollCalls(), 0)
			}
		}).Repeat(20))
	}
}

func randomEvenShareCounts() map[string]int64 {
	shareCounts := make(map[string]int64)

	total := sdk.ZeroInt()
	for i := 0; i < int(rand.I64Between(1, 20)); i++ {
		addr := sdk.ValAddress(rand.Bytes(20))
		share := rand.PosI64()
		shareCounts[addr.String()] = share
		total = total.AddRaw(share)
	}

	// redraw shares if any one share is greater than half of the total
	for _, share := range shareCounts {
		if total.QuoRaw(2).LT(sdk.NewInt(share)) {
			return randomEvenShareCounts()
		}
	}

	return shareCounts
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
	poll := types.NewPollMetaData(key, types.DefaultGenesisState().VotingThreshold, rand.PosI64())
	poll.ExpiresAt = rand.I64Between(0, 1000000)
	return poll
}
