package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
)

func TestMaintainerState(t *testing.T) {
	t.Run("MarkIncorrectVote and CountIncorrectVotes", func(t *testing.T) {
		ms := NewMaintainerState(testutils.Chain().Name, rand.ValAddr())

		ms.MarkIncorrectVote(true)
		ms.MarkIncorrectVote(true)
		ms.MarkIncorrectVote(false)

		assert.EqualValues(t, 2, ms.CountIncorrectVotes(10))
		assert.EqualValues(t, 0, ms.CountMissingVotes(10))
	})

	t.Run("MarkMissingVote and CountMissingVotes", func(t *testing.T) {
		ms := NewMaintainerState(testutils.Chain().Name, rand.ValAddr())

		ms.MarkMissingVote(true)
		ms.MarkMissingVote(false)

		assert.EqualValues(t, 0, ms.CountIncorrectVotes(10))
		assert.EqualValues(t, 1, ms.CountMissingVotes(10))
	})
}
