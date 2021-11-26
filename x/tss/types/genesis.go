package types

import (
	"encoding/json"
	fmt "fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState is the constructor for GenesisState
func NewGenesisState(
	params Params,
	governanceKey *multisig.LegacyAminoPubKey,
	keyRecoveryInfos []KeyRecoveryInfo,
	keys []exported.Key,
	multisigInfos []MultisigInfo,
	externalKeys []ExternalKeys,
) *GenesisState {
	return &GenesisState{
		Params:           params,
		GovernanceKey:    governanceKey,
		KeyRecoveryInfos: keyRecoveryInfos,
		Keys:             keys,
		MultisigInfos:    multisigInfos,
		ExternalKeys:     externalKeys,
	}
}

// DefaultGenesis represents the default genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		nil,
		[]KeyRecoveryInfo{},
		[]exported.Key{},
		[]MultisigInfo{},
		[]ExternalKeys{},
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

	if m.GovernanceKey != nil {
		if int(m.GovernanceKey.Threshold) > len(m.GovernanceKey.PubKeys) {
			return fmt.Errorf("threshold k of n multisignature: len(pubKeys) < k")
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
