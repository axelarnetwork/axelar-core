package types

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	// TSSEnabled indicates if threshold signing is permitted
	TSSEnabled bool = false
)

// Validate validates the ValidatorStatus
func (m ValidatorStatus) Validate() error {
	if err := sdk.VerifyAddressFormat(m.Validator); err != nil {
		return err
	}

	return nil
}

// Validate validates the MultisigInfo
func (m MultisigInfo) Validate() error {
	if err := utils.ValidateString(m.ID); err != nil {
		return sdkerrors.Wrap(err, "invalid ID")
	}

	if m.Timeout <= 0 {
		return fmt.Errorf("timeout must be >0")
	}

	if m.TargetNum <= 0 {
		return fmt.Errorf("target num must be >0")
	}

	return nil
}

// Validate validates the ExternalKeys
func (m ExternalKeys) Validate() error {
	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if len(m.KeyIDs) == 0 {
		return fmt.Errorf("key IDs must be set")
	}

	for _, keyID := range m.KeyIDs {
		if err := keyID.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the KeyRecoveryInfo
func (m KeyRecoveryInfo) Validate() error {
	if err := m.KeyID.Validate(); err != nil {
		return err
	}

	if len(m.Public) == 0 {
		return fmt.Errorf("missing public key recovery info")
	}

	if len(m.Private) == 0 {
		return fmt.Errorf("missing private key recovery info")
	}

	for validator, privateKeyRecoveryInfo := range m.Private {
		if len(privateKeyRecoveryInfo) == 0 {
			return fmt.Errorf("missing private key recovery info for validator %s", validator)
		}
	}

	return nil
}

// MultisigBaseInfo is an interface for multisig base info
type MultisigBaseInfo interface {
	HasData(k []byte) bool
	AddData(val sdk.ValAddress, data [][]byte)
	GetData() [][]byte
	IsCompleted() bool
	Count() int64
	DoesParticipate(val sdk.ValAddress) bool
	GetTimeoutBlock() int64
}

// MultisigKeygenInfo is an interface for multisig keygen info
type MultisigKeygenInfo interface {
	MultisigBaseInfo
	GetKeys() []ecdsa.PublicKey
	GetPubKeysByValidator(val sdk.ValAddress) []ecdsa.PublicKey
}

// MultisigSignInfo is an interface for multisig sign info
type MultisigSignInfo interface {
	MultisigBaseInfo
	GetTargetSigKeyPairs() []exported.SigKeyPair
}

// HasData checks duplicate data
func (m MultisigInfo) HasData(k []byte) bool {
	for _, info := range m.Infos {
		for _, d := range info.Data {
			if res := bytes.Compare(k, d); res == 0 {
				return true
			}
		}
	}

	return false
}

// AddData add list of data for a participant
func (m *MultisigInfo) AddData(val sdk.ValAddress, data [][]byte) {
	for _, info := range m.Infos {
		if val.Equals(info.Participant) {
			info.Data = append(info.Data, data...)
			return
		}
	}
	m.Infos = append(m.Infos, &MultisigInfo_Info{Participant: val, Data: data})
}

// GetData returns list of data
func (m *MultisigInfo) GetData() [][]byte {
	var data [][]byte
	for _, info := range m.Infos {
		data = append(data, info.Data...)
	}

	return data
}

// IsCompleted returns true if number of data reaches target
func (m MultisigInfo) IsCompleted() bool {
	return m.Count() >= m.TargetNum
}

// Count returns current number of data
func (m MultisigInfo) Count() int64 {
	l := 0
	for _, info := range m.Infos {
		l += len(info.Data)
	}
	return int64(l)
}

// DoesParticipate returns true if the validator submitted its data
func (m MultisigInfo) DoesParticipate(val sdk.ValAddress) bool {
	for _, info := range m.Infos {
		if val.Equals(info.Participant) {
			return true
		}
	}
	return false
}

// GetTimeoutBlock returns multisig info timeout height
func (m MultisigInfo) GetTimeoutBlock() int64 {
	return m.Timeout
}

// GetKeys returns list of all pub keys
func (m MultisigInfo) GetKeys() []ecdsa.PublicKey {
	var pubKeys []ecdsa.PublicKey
	for _, info := range m.Infos {
		for _, pubKey := range info.Data {
			btcecPK, err := btcec.ParsePubKey(pubKey, btcec.S256())
			// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
			if err != nil {
				panic(err)
			}
			pk := btcecPK.ToECDSA()
			pubKeys = append(pubKeys, *pk)
		}
	}

	return pubKeys
}

// GetPubKeysByValidator returns pub keys a validator submitted
func (m MultisigInfo) GetPubKeysByValidator(val sdk.ValAddress) []ecdsa.PublicKey {
	var pubKeys []ecdsa.PublicKey

	for _, info := range m.Infos {
		if val.Equals(info.Participant) {
			for _, pubKey := range info.Data {
				btcecPK, err := btcec.ParsePubKey(pubKey, btcec.S256())
				// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
				if err != nil {
					panic(err)
				}
				pk := btcecPK.ToECDSA()
				pubKeys = append(pubKeys, *pk)
			}
		}
	}

	return pubKeys
}

// GetTargetSigKeyPairs returns list of pub key and signature pairs
func (m MultisigInfo) GetTargetSigKeyPairs() []exported.SigKeyPair {
	var pairs []exported.SigKeyPair
	for _, info := range m.Infos {
		for _, sigKeyPair := range info.Data {
			var pair exported.SigKeyPair

			err := pair.Unmarshal(sigKeyPair)
			if err != nil {
				panic(err)
			}
			pairs = append(pairs, pair)
			if len(pairs) >= int(m.TargetNum) {
				return pairs
			}
		}
	}

	return pairs
}
