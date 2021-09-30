package exported_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

func TestComputeAbsCorruptionThreshold(t *testing.T) {
	assert.Equal(t, int64(7), exported.ComputeAbsCorruptionThreshold(utils.NewThreshold(2, 3), sdk.NewInt(12)))
	assert.Equal(t, int64(3), exported.ComputeAbsCorruptionThreshold(utils.NewThreshold(11, 20), sdk.NewInt(7)))
}
