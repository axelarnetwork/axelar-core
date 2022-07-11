package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
)

func TestDefaultGenesisState(t *testing.T) {
	assert.NoError(t, types.DefaultGenesisState().Validate())
}
