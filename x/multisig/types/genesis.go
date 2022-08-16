package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

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

	if err := validateKeys(m.KeygenSessions, m.Keys); err != nil {
		return getValidateError(err)
	}

	keys := slices.ToMap(m.Keys, func(key Key) exported.KeyID { return key.ID })

	if err := validateSigningSessions(keys, m.SigningSessions); err != nil {
		return getValidateError(err)
	}

	if err := validateKeyEpochs(keys, m.KeyEpochs); err != nil {
		return getValidateError(err)
	}

	return nil
}

func validateKeyEpochs(keys map[exported.KeyID]Key, keyEpochs []KeyEpoch) error {
	keyIDSeen := make(map[string]bool, len(keyEpochs))
	for _, kes := range slices.GroupBy(keyEpochs, func(keyEpoch KeyEpoch) string { return keyEpoch.GetChain().String() }) {
		sort.SliceStable(kes, func(i, j int) bool { return kes[i].Epoch < kes[j].Epoch })

		for i, keyEpoch := range kes {
			if keyEpoch.Epoch != uint64(i+1) {
				return fmt.Errorf("invalid epoch set for key epoch")
			}

			keyIDLowerCase := strings.ToLower(keyEpoch.GetKeyID().String())
			if keyIDSeen[keyIDLowerCase] {
				return fmt.Errorf("duplicate key ID seen in key epochs")
			}
			keyIDSeen[keyIDLowerCase] = true

			if _, ok := keys[keyEpoch.GetKeyID()]; !ok {
				return fmt.Errorf("key ID %s in key epoch does not exist", keyEpoch.GetKeyID())
			}

			if err := keyEpoch.ValidateBasic(); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSigningSessions(keys map[exported.KeyID]Key, signingSessions []SigningSession) error {
	sigIDSeen := make(map[uint64]bool, len(signingSessions))
	for _, signingSession := range signingSessions {
		if sigIDSeen[signingSession.GetID()] {
			return fmt.Errorf("duplicate key ID seen")
		}
		sigIDSeen[signingSession.GetID()] = true

		if _, ok := keys[signingSession.Key.ID]; !ok {
			return fmt.Errorf("key ID %s in signature does not exist", signingSession.Key.ID)
		}

		if err := signingSession.ValidateBasic(); err != nil {
			return err
		}
	}
	return nil
}

func validateKeys(keygenSessions []KeygenSession, keys []Key) error {
	keyIDSeen := make(map[string]bool, len(keygenSessions)+len(keys))
	for _, keygenSession := range keygenSessions {
		keyIDLowerCase := strings.ToLower(keygenSession.GetKeyID().String())
		if keyIDSeen[keyIDLowerCase] {
			return fmt.Errorf("duplicate key ID seen in keygen sessions")
		}
		keyIDSeen[keyIDLowerCase] = true

		if err := keygenSession.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, key := range keys {
		keyIDLowerCase := strings.ToLower(key.GetID().String())
		if keyIDSeen[keyIDLowerCase] {
			return fmt.Errorf("duplicate key ID seen in keys")
		}
		keyIDSeen[keyIDLowerCase] = true

		if err := key.ValidateBasic(); err != nil {
			return err
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
