package vote

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

// GetGenesisStateFromAppState returns x/vote GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) types.GenesisState {
	var genesisState types.GenesisState
	if appState[types.ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisState)
	}

	return genesisState
}
