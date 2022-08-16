package types

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/utils/slices"
)

// NewGenesisState is the constructor for GenesisState
func NewGenesisState(params Params, keygenSessions []KeygenSession, signingSessions []SigningSession, keys []Key, keyEpochs []KeyEpoch) *GenesisState {
	return &GenesisState{
		Params:          params,
		KeygenSessions:  keygenSessions,
		SigningSessions: signingSessions,
		Keys:            keys,
		KeyEpochs:       keyEpochs,
	}
}

// DefaultGenesisState returns a genesis state with default parameters
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]KeygenSession{},
		[]SigningSession{},
		[]Key{},
		[]KeyEpoch{},
	)
}

// Validate performs a validation check on the genesis parameters
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return getValidateError(err)
	}

	keyIDs := make(map[exported.KeyID]bool, len(m.KeygenSessions)+len(m.Keys))
	completedKeyIDs := make(map[exported.KeyID]bool)
	for _, keygenSession := range m.KeygenSessions {
		if keyIDs[keygenSession.GetKeyID()] {
			return getValidateError(fmt.Errorf("duplicate key ID seen"))
		}
		keyIDs[keygenSession.GetKeyID()] = true

		if err := keygenSession.ValidateBasic(); err != nil {
			return getValidateError(err)
		}
	}
	for _, key := range m.Keys {
		if keyIDs[key.GetID()] {
			return getValidateError(fmt.Errorf("duplicate key ID seen"))
		}
		keyIDs[key.GetID()] = true
		completedKeyIDs[key.GetID()] = true

		if err := key.ValidateBasic(); err != nil {
			return getValidateError(err)
		}
	}

	sigIDs := make(map[uint64]bool, len(m.SigningSessions))
	for _, signingSession := range m.SigningSessions {
		if sigIDs[signingSession.GetID()] {
			return getValidateError(fmt.Errorf("duplicate key ID seen"))
		}
		sigIDs[signingSession.GetID()] = true

		if !completedKeyIDs[signingSession.Key.ID] {
			return getValidateError(fmt.Errorf("key ID %s in signature does not exist", signingSession.Key.ID))
		}

		if err := signingSession.ValidateBasic(); err != nil {
			return getValidateError(err)
		}
	}

	for _, keyEpochs := range slices.GroupBy(m.KeyEpochs, func(keyEpoch KeyEpoch) string { return keyEpoch.GetChain().String() }) {
		sort.SliceStable(keyEpochs, func(i, j int) bool { return keyEpochs[i].Epoch < keyEpochs[j].Epoch })

		for i, keyEpoch := range keyEpochs {
			if keyEpoch.Epoch != uint64(i+1) {
				return getValidateError(fmt.Errorf("invalid epoch set for key epoch"))
			}

			if err := keyEpoch.ValidateBasic(); err != nil {
				return getValidateError(err)
			}
		}
	}

	return nil
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}

// GetGenesisStateFromAppState returns the GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
