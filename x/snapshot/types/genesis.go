package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GenesisState - genesis state of the snapshot module
type GenesisState struct {
	Params Params
}

// DefaultGenesisState returns a genesis state with default parameters
func DefaultGenesisState() GenesisState {
	return GenesisState{Params: DefaultParams()}
}

// ValidateGenesis performs a validation check on the genesis parameters
func ValidateGenesis(state GenesisState) error {
	if err := state.Params.Validate(); err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("genesis state for module %s is invalid", ModuleName))
	}

	return nil
}

// GetGenesisStateFromAppState returns x/snapshot GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
