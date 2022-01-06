package exported_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func TestPollKey_Validate(t *testing.T) {
	t.Run("empty module", func(t *testing.T) {
		assert.Error(t, exported.NewPollKey("", randomNormalizedStr(10, 10)).Validate())
	})

	t.Run("empty id", func(t *testing.T) {
		assert.Error(t, exported.NewPollKey(randomNormalizedStr(10, 10), "").Validate())
	})

	t.Run("correct key", testutils.Func(func(t *testing.T) {
		id := randomNormalizedStr(5, 20)
		module := randomNormalizedStr(5, 20)
		assert.NoError(t, exported.NewPollKey(module, id).Validate())
	}).Repeat(20))
}

func randomNormalizedStr(min, max int) string {
	return strings.ReplaceAll(utils.NormalizeString(rand.StrBetween(min, max)), utils.DefaultDelimiter, "-")
}
