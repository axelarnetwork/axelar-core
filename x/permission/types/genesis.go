package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// NewGenesisState is the constructor for GenesisState
func NewGenesisState(params Params, governanceKey *multisig.LegacyAminoPubKey, accounts []GovAccount) *GenesisState {
	return &GenesisState{
		Params:        params,
		GovernanceKey: governanceKey,
		GovAccounts:   accounts,
	}
}

// DefaultGenesisState returns a genesis state with default parameters
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(Params{}, nil, []GovAccount{})
}

// Validate performs a validation check on the genesis parameters
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return getValidateError(err)
	}

	if m.GovernanceKey != nil {
		if int(m.GovernanceKey.Threshold) > len(m.GovernanceKey.PubKeys) {
			return fmt.Errorf("threshold k of n multisignature: len(pubKeys) < k")
		}
	}

	accessControlSet := false
	for _, account := range m.GovAccounts {
		if err := account.Validate(); err != nil {
			return getValidateError(err)
		}

		// exactly one account with role ROLE_ACCESS_CONTROL
		if account.Role == exported.ROLE_ACCESS_CONTROL {
			if accessControlSet {
				return getValidateError(fmt.Errorf("role access control already set"))
			}
			accessControlSet = true
		}
	}

	return nil
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}

// GetGenesisStateFromAppState returns x/permission GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
