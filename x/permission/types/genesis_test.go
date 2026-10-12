package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

func TestDefaultGenesisState(t *testing.T) {
	assert.NoError(t, DefaultGenesisState().Validate())
}

func TestGenesisState_ValidatePauseLevel(t *testing.T) {
	for _, level := range []exported.PauseLevel{exported.PAUSE_LEVEL_UNSPECIFIED, exported.PAUSE_LEVEL_ALL, exported.PAUSE_LEVEL_EXCEPT_VALIDATORS} {
		genState := DefaultGenesisState()
		genState.PauseLevel = level
		assert.NoError(t, genState.Validate())
	}

	genState := DefaultGenesisState()
	genState.PauseLevel = exported.PauseLevel(1000)
	assert.Error(t, genState.Validate())
}
