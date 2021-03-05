package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type GenesisState struct {
	Params Params
}

func DefaultGenesisState() GenesisState {
	return GenesisState{DefaultParams()}
}

func ValidateGenesis(g GenesisState) error {
	if err := g.Params.Validate(); err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("genesis state for module %s is invalid", ModuleName))
	}
	return nil
}

// GetGenesisStateFromAppState returns x/tss GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
