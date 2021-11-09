package types

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// TSSEnabled indicates if threshold signing is permitted
	TSSEnabled bool = false
)

// MultisigBaseInfo is an interface for multisig base info
type MultisigBaseInfo interface {
	HasData(k []byte) bool
	AddData(val sdk.ValAddress, data [][]byte)
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
	GetSigs() []exported.Signature
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

// GetSigs returns list of all signatures
func (m MultisigInfo) GetSigs() []exported.Signature {
	var signatures []exported.Signature
	for _, info := range m.Infos {
		for _, sig := range info.Data {
			btcecSig, err := btcec.ParseDERSignature(sig, btcec.S256())
			// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
			if err != nil {
				panic(err)
			}

			signatures = append(signatures, exported.Signature{R: btcecSig.R, S: btcecSig.S})
		}
	}

	return signatures
}
