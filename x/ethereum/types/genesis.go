package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type GenesisState struct {
	Params Params
}

func DefaultGenesisState() GenesisState {
	return GenesisState{DefaultParams()}
}

func ValidateGenesis(state GenesisState) error {
	if err := state.Params.Validate(); err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("genesis state for module %s is invalid", ModuleName))
	}

	return nil
}
