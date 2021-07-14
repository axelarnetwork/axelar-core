package exported_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

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
