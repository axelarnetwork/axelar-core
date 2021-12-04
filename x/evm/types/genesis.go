package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{Chains: DefaultChains()}
}

func DefaultChains() []GenesisState_Chain {
	p := DefaultParams()

	var chains []GenesisState_Chain
	for _, params := range p {
		chain := GenesisState_Chain{
			Params:            params,
			BurnerInfos:       nil,
			CommandQueue:      nil,
			ConfirmedDeposits: nil,
			BurnedDeposits:    nil,
			LatestBatch:       CommandBatchMetadata{},
			SignedBatches:     nil,
			Gateway:           Gateway{},
			Tokens:            nil,
		}
		chains = append(chains, chain)
	}
	return chains
}

// Validate calidates the genesis state
func (m GenesisState) Validate() error {
	for _, chain := range m.Chains {
		if err := chain.Params.Validate(); err != nil {
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
