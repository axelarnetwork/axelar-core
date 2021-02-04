package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GenesisState represents the genesis state
type GenesisState struct {
	Params Params
}

// DefaultGenesisState creates the default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{DefaultParams()}
}

// ValidateGenesis checks if the genesis state is valid
func ValidateGenesis(state GenesisState) error {
	if err := state.Params.Validate(); err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("genesis state for module %s is invalid", ModuleName))
	}

	return nil
}
