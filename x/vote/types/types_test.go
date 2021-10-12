package types_test

import (
	"fmt"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
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

func TestPoll_Expiry(t *testing.T) {
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
		currBlockHeight := expiry - rand.I64Between(0, expiry)

		poll := types.NewPoll(metadata, currBlockHeight, &mock.StoreMock{})

		assert.True(t, poll.Is(initialState))
	}).Repeat(repeats))

	t.Run("pending poll expires", testutils.Func(func(t *testing.T) {
		metadata := setup()
		metadata.State = exported.Pending
		expiry := rand.I64Between(0, 1000000)
		metadata.ExpiresAt = expiry
		currBlockHeight := expiry + rand.I64Between(1, 1000000)

		poll := types.NewPoll(metadata, currBlockHeight, &mock.StoreMock{})

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
		currBlockHeight := expiry + rand.I64Between(1, 1000000)
		poll := types.NewPoll(metadata, currBlockHeight, &mock.StoreMock{})

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

	setup := func(metadata exported.PollMetadata, currBlockHeight int64) *types.Poll {
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

		return types.NewPoll(metadata, currBlockHeight, store).WithLogger(log.TestingLogger())
	}
	repeats := 20

	t.Run("poll nonexistent", testutils.Func(func(t *testing.T) {
		poll := setup(exported.PollMetadata{State: exported.NonExistent}, rand.PosI64())

		voter := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voter.String()] = rand.PosI64()
		totalShareCount = sdk.NewInt(shareCounts[voter.String()]).MulRaw(10)

		assert.Error(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("poll already completed", testutils.Func(func(t *testing.T) {
		result, _ := codectypes.NewAnyWithValue(&gogoprototypes.StringValue{Value: rand.Str(10)})
		poll := setup(exported.PollMetadata{State: exported.Completed, Result: result}, rand.PosI64())

		voter := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voter.String()] = rand.PosI64()
		totalShareCount = sdk.NewInt(shareCounts[voter.String()]).MulRaw(10)

		assert.NoError(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("poll already failed", testutils.Func(func(t *testing.T) {
		poll := setup(exported.PollMetadata{State: exported.Failed}, rand.PosI64())

		voter := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		shareCounts[voter.String()] = rand.PosI64()
		totalShareCount = sdk.NewInt(shareCounts[voter.String()]).MulRaw(10)

		assert.NoError(t, poll.Vote(voter, &gogoprototypes.StringValue{Value: rand.Str(10)}))
	}).Repeat(repeats))

	t.Run("voter unknown", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())

		voterAddr := sdk.ValAddress(rand.Bytes(sdk.AddrLen))
		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}

		assert.Error(t, poll.Vote(voterAddr, voteValue))
	}).Repeat(repeats))

	t.Run("correct vote no completion", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())

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
		currBlockHeight := metadata.ExpiresAt + rand.I64Between(1, 1000000)
		poll := setup(metadata, currBlockHeight)

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
		poll := setup(metadata, rand.PosI64())

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
		poll := setup(metadata, rand.PosI64())

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
		poll := setup(metadata, rand.PosI64())

		t.Log(len(shareCounts), "votes")
		for voter := range shareCounts {
			addr, _ := sdk.ValAddressFromBech32(voter)
			voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
			assert.NoError(t, poll.Vote(addr, voteValue))
		}

		assert.True(t, poll.Is(exported.Failed), poll.State)
	}).Repeat(repeats))

	t.Run("should complete the poll when total voter count is less than minimum voter count and minimum voter count is not met", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())
		poll.MinVoterCount = int64(len(shareCounts) + 1)

		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
		for voter := range shareCounts {
			addr, _ := sdk.ValAddressFromBech32(voter)
			assert.NoError(t, poll.Vote(addr, voteValue))
		}

		assert.True(t, poll.Is(exported.Completed))
		assert.Equal(t, voteValue, poll.GetResult())
	}).Repeat(repeats))

	t.Run("should not complete the poll when total voter count is greater than or equal to minimum voter count and minimum voter count is not met", testutils.Func(func(t *testing.T) {
		metadata := newRandomPollMetadata()
		poll := setup(metadata, rand.PosI64())
		poll.MinVoterCount = int64(len(shareCounts))
		poll.VotingThreshold = utils.ZeroThreshold

		voteValue := &gogoprototypes.StringValue{Value: rand.StrBetween(1, 500)}
		voterCount := int64(0)
		for voter := range shareCounts {
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
			poll := types.NewPoll(newRandomPollMetadata(), rand.PosI64(), store).WithLogger(log.TestingLogger())

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
		return types.NewPoll(metadata, rand.I64Between(0, metadata.ExpiresAt), store).WithLogger(log.TestingLogger())
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

func randomEvenShareCounts() map[string]int64 {
	shareCounts := make(map[string]int64)

	total := sdk.ZeroInt()
	for i := 0; i < int(rand.I64Between(1, 20)); i++ {
		addr := sdk.ValAddress(rand.Bytes(20))
		share := rand.I64Between(1, 100)
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
	poll.ExpiresAt = rand.I64Between(1, 1000000)
	return poll
}
