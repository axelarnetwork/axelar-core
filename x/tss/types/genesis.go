package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewGenesisState is the constructor for GenesisState
func NewGenesisState(
	params Params,
	keyRecoveryInfos []KeyRecoveryInfo,
	keys []exported.Key,
	multisigInfos []MultisigInfo,
	externalKeys []ExternalKeys,
	signatures []exported.Signature,
	validatorStatuses []ValidatorStatus,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		KeyRecoveryInfos:  keyRecoveryInfos,
		Keys:              keys,
		MultisigInfos:     multisigInfos,
		ExternalKeys:      externalKeys,
		Signatures:        signatures,
		ValidatorStatuses: validatorStatuses,
	}
}

// DefaultGenesisState represents the default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]KeyRecoveryInfo{},
		[]exported.Key{},
		[]MultisigInfo{},
		[]ExternalKeys{},
		[]exported.Signature{},
		[]ValidatorStatus{},
	)
}

// Validate validates the genesis state
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return getValidateError(err)
	}

	for _, keyRecoveryInfo := range m.KeyRecoveryInfos {
		if err := keyRecoveryInfo.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, key := range m.Keys {
		if err := key.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, multisigInfo := range m.MultisigInfos {
		if err := multisigInfo.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, externalKeys := range m.ExternalKeys {
		if err := externalKeys.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, signature := range m.Signatures {
		if err := signature.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, validatorStatus := range m.ValidatorStatuses {
		if err := validatorStatus.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	return nil
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}

// GetGenesisStateFromAppState returns x/tss GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
