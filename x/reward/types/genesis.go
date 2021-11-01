package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesisState returns a genesis state with default parameters
func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}

// Validate performs a validation check on the genesis parameters
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("genesis state for module %s is invalid", ModuleName))
	}

	return nil
}
