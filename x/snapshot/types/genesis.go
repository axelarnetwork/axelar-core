package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

// NewGenesisState is the constructor for GenesisState
func NewGenesisState(params Params, snapshots []exported.Snapshot, proxiedValidators []ProxiedValidator) *GenesisState {
	return &GenesisState{
		Params:            params,
		Snapshots:         snapshots,
		ProxiedValidators: proxiedValidators,
	}
}

// DefaultGenesisState returns a genesis state with default parameters
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []exported.Snapshot{}, []ProxiedValidator{})
}

// Validate performs a validation check on the genesis parameters
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return getValidateError(err)
	}

	for i, snapshot := range m.Snapshots {
		if snapshot.Counter != int64(i) {
			return getValidateError(fmt.Errorf("snapshot counter has to be sequential"))
		}

		if err := snapshot.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, proxiedValidator := range m.ProxiedValidators {
		if err := proxiedValidator.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	return nil
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}

// GetGenesisStateFromAppState returns x/snapshot GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
