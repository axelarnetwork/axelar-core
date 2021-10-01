package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{DefaultParams()}
}

// Validate calidates the genesis state
func (m GenesisState) Validate() error {
	for _, p := range m.Params {
		if err := p.Validate(); err != nil {
			return sdkerrors.Wrap(err, fmt.Sprintf("genesis m for module %s is invalid", ModuleName))
		}
	}

	return nil
}

// GetGenesisStateFromAppState returns x/evm GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
