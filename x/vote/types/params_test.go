package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/testutils"
)

func TestParams_Validate(t *testing.T) {
	t.Run("zero end blocker limit", func(t *testing.T) {
		params := Params{
			DefaultVotingThreshold: testutils.RandThreshold(),
			EndBlockerLimit:        int64(0),
		}
		assert.Error(t, params.Validate())
	})

	t.Run("negative end blocker limit", func(t *testing.T) {
		params := Params{
			DefaultVotingThreshold: testutils.RandThreshold(),
			EndBlockerLimit:        -rand.PosI64(),
		}
		assert.Error(t, params.Validate())
	})

	t.Run("zero threshold", func(t *testing.T) {
		params := Params{
			DefaultVotingThreshold: utils.ZeroThreshold,
			EndBlockerLimit:        rand.PosI64(),
		}
		assert.Error(t, params.Validate())
	})

	t.Run("negative threshold", func(t *testing.T) {
		threshold := testutils.RandThreshold()
		threshold.Numerator *= -1
		params := Params{
			DefaultVotingThreshold: threshold,
			EndBlockerLimit:        rand.PosI64(),
		}
		assert.Error(t, params.Validate())
	})

	t.Run("greater than one threshold", func(t *testing.T) {
		threshold := utils.OneThreshold
		threshold.Numerator += threshold.Denominator
		params := Params{
			DefaultVotingThreshold: threshold,
			EndBlockerLimit:        rand.PosI64(),
		}
		assert.Error(t, params.Validate())
	})

	t.Run("correct params", func(t *testing.T) {
		params := Params{
			DefaultVotingThreshold: testutils.RandThreshold(),
			EndBlockerLimit:        rand.PosI64(),
		}
		assert.NoError(t, params.Validate())
	})
}
