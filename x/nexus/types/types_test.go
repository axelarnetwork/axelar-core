package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
)

func TestChainState(t *testing.T) {
	t.Run("pad false votes in range", func(t *testing.T) {
		state := ChainState{
			Chain:            testutils.Chain(),
			Activated:        false,
			Assets:           nil,
			MaintainerStates: nil,
		}

		maintainer := rand.ValAddr()
		assert.NoError(t, state.AddMaintainer(maintainer))
		state.MarkIncorrectVote(maintainer, true)
		state.MarkIncorrectVote(maintainer, true)
		state.MarkIncorrectVote(maintainer, false)

		state.MarkMissingVote(maintainer, true)
		state.MarkMissingVote(maintainer, false)

		assert.Len(t, state.MaintainerStates, 1)

		assert.EqualValues(t, 2, state.MaintainerStates[0].IncorrectVotes.CountTrue(10))
		assert.EqualValues(t, 8, state.MaintainerStates[0].IncorrectVotes.CountFalse(10))
		assert.EqualValues(t, 1, state.MaintainerStates[0].MissingVotes.CountTrue(10))
		assert.EqualValues(t, 9, state.MaintainerStates[0].MissingVotes.CountFalse(10))
	})

	t.Run("do not record vote state for unknown maintainer", func(t *testing.T) {
		state := ChainState{
			Chain:            testutils.Chain(),
			Activated:        false,
			Assets:           nil,
			MaintainerStates: nil,
		}

		state.MarkIncorrectVote(rand.ValAddr(), true)
		state.MarkIncorrectVote(rand.ValAddr(), false)

		state.MarkMissingVote(rand.ValAddr(), true)
		state.MarkMissingVote(rand.ValAddr(), false)

		assert.Len(t, state.MaintainerStates, 0)
	})
}
