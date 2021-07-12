package exported_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

func TestPollMetadata_Is(t *testing.T) {
	for _, state := range exported.PollState_value {
		key := exported.NewPollKey(rand.StrBetween(5, 20), rand.StrBetween(5, 20))
		expiresAt := rand.I64Between(0, 1000000)
		poll := exported.NewPollMetaData(key, rand.PosI64(), expiresAt, types.DefaultGenesisState().VotingThreshold)

		poll.State = exported.PollState(state)

		assert.True(t, poll.Is(exported.PollState(state)))

		for _, otherState := range exported.PollState_value {
			if otherState == state {
				continue
			}
			assert.False(t, poll.Is(exported.PollState(otherState)), "poll: %s, other: %s", poll.State, exported.PollState(otherState))
		}
	}
}

func TestPollMetadata_UpdateBlockHeight(t *testing.T) {
	setup := func() exported.PollMetadata {
		key := exported.NewPollKey(rand.StrBetween(5, 20), rand.StrBetween(5, 20))
		expiresAt := rand.I64Between(0, 1000000)
		return exported.NewPollMetaData(key, rand.PosI64(), expiresAt, types.DefaultGenesisState().VotingThreshold)
	}
	repeats := 20

	states := []exported.PollState{exported.NonExistent, exported.Pending, exported.Completed, exported.Failed}

	t.Run("poll is not expired", testutils.Func(func(t *testing.T) {
		poll := setup()
		initialState := states[rand.I64Between(0, int64(len(states)))]
		poll.State = initialState

		assert.True(t, poll.Is(initialState))

		poll = poll.UpdateBlockHeight(poll.ExpiresAt - rand.I64Between(1, poll.ExpiresAt))
		assert.True(t, poll.Is(initialState))
	}).Repeat(repeats))

	t.Run("pending poll expires", testutils.Func(func(t *testing.T) {
		poll := setup()
		poll.State = exported.Pending

		assert.True(t, poll.Is(exported.Pending))

		poll = poll.UpdateBlockHeight(poll.ExpiresAt + rand.I64Between(0, 1000000))
		assert.True(t, poll.Is(exported.Pending))
		assert.True(t, poll.Is(exported.Expired))
	}).Repeat(repeats))

	t.Run("not-pending poll does not expire", testutils.Func(func(t *testing.T) {
		poll := setup()

		stateIdx := rand.I64GenBetween(0, int64(len(states))).
			Where(func(i int64) bool { return i != int64(exported.Pending) }).Next()
		initialState := states[stateIdx]
		poll.State = initialState

		assert.True(t, poll.Is(initialState))

		poll = poll.UpdateBlockHeight(poll.ExpiresAt + rand.I64Between(0, 1000000))
		assert.True(t, poll.Is(initialState))
		assert.False(t, poll.Is(exported.Expired))
	}).Repeat(repeats))
}

func TestPollKey_Validate(t *testing.T) {
	t.Run("empty module", func(t *testing.T) {
		assert.Error(t, exported.NewPollKey("", rand.Str(10)).Validate())
	})

	t.Run("empty id", func(t *testing.T) {
		assert.Error(t, exported.NewPollKey(rand.Str(10), "").Validate())
	})

	t.Run("correct key", testutils.Func(func(t *testing.T) {
		id := rand.StrBetween(5, 20)
		module := rand.StrBetween(5, 20)
		assert.NoError(t, exported.NewPollKey(module, id).Validate())
	}).Repeat(20))
}
