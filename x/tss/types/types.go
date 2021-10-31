package types

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HasKey checks duplicate pub keys
func (m MultisigKeyInfo) HasKey(k []byte) bool {
	for _, pubKey := range m.PubKeys {
		if res := bytes.Compare(k, pubKey); res == 0 {
			return true
		}
	}
	return false
}

// AddKey append pub key to PubKeys list
func (m *MultisigKeyInfo) AddKey(k []byte) {
	m.PubKeys = append(m.PubKeys, k)
}

// AddParticipant stores the validator who submitted its pub keys
func (m *MultisigKeyInfo) AddParticipant(val sdk.ValAddress) {
	m.Participants = append(m.Participants, val)
}

// IsCompleted returns true if multisig keygen complete
func (m *MultisigKeyInfo) IsCompleted() bool {
	return int64(len(m.PubKeys)) == m.TargetKeyNum
}

// GetKeys returns list of pub keys
func (m MultisigKeyInfo) GetKeys() []ecdsa.PublicKey {
	var pubKeys []ecdsa.PublicKey
	for _, pubKey := range m.PubKeys {
		btcecPK, err := btcec.ParsePubKey(pubKey, btcec.S256())
		// the setter is controlled by the keeper alone, so an error here should be a catastrophic failure
		if err != nil {
			panic(err)
		}
		pk := btcecPK.ToECDSA()
		pubKeys = append(pubKeys, *pk)
	}

	return pubKeys
}

// KeyCount returns current number of pub keys
func (m *MultisigKeyInfo) KeyCount() int64 {
	return int64(len(m.PubKeys))
}

// DoesParticipate returns true if the validator submitted its pub keys
func (m MultisigKeyInfo) DoesParticipate(val sdk.ValAddress) bool {
	for _, p := range m.Participants {
		if val.Equals(p) {
			return true
		}
	}
	return false
}
