package types

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

const (
	ModeSpecificAddress Mode = iota
	ModeCurrentMasterKey
	ModeSpecificKey
)

type Mode int

// OutPointInfo describes all the necessary information to verify the outPoint of a transaction
type OutPointInfo struct {
	OutPoint      *wire.OutPoint
	Amount        btcutil.Amount
	Recipient     BtcAddress
	Confirmations uint64
}

// Validate ensures that all fields are filled with sensible values
func (u OutPointInfo) Validate() error {
	if u.OutPoint == nil {
		return fmt.Errorf("missing outpoint")
	}
	if u.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if err := u.Recipient.Validate(); err != nil {
		return err
	}
	return nil
}

// Equals checks if two OutPointInfo objects are semantically equal
func (u OutPointInfo) Equals(other OutPointInfo) bool {
	return u.OutPoint.Hash.IsEqual(&other.OutPoint.Hash) &&
		u.OutPoint.Index == other.OutPoint.Index &&
		u.Amount == other.Amount &&
		u.Recipient == other.Recipient
}

// Network provides additional functionality based on the bitcoin network name
type Network string

// Validate checks if the object is a valid chain
func (c Network) Validate() error {
	switch string(c) {
	case chaincfg.MainNetParams.Name, chaincfg.TestNet3Params.Name, chaincfg.RegressionNetParams.Name:
		return nil
	default:
		return fmt.Errorf("chain could not be parsed, choose %s, %s, or %s",
			chaincfg.MainNetParams.Name,
			chaincfg.TestNet3Params.Name,
			chaincfg.RegressionNetParams.Name,
		)
	}
}

// Params returns the configuration parameters associated with the chain
func (c Network) Params() *chaincfg.Params {
	switch string(c) {
	case chaincfg.MainNetParams.Name:
		return &chaincfg.MainNetParams
	case chaincfg.TestNet3Params.Name:
		return &chaincfg.TestNet3Params
	case chaincfg.RegressionNetParams.Name:
		return &chaincfg.RegressionNetParams
	default:
		return nil
	}
}

// PKHashFromKey creates a Bitcoin pubKey hash address from a public key.
// We use Pay2PKH for added security over Pay2PK as well as for the benefit of getting a parsed address from the response of
// getrawtransaction() on the Bitcoin rpc client
func PKHashFromKey(key ecdsa.PublicKey, chain Network) (*btcutil.AddressPubKeyHash, error) {
	btcPK := btcec.PublicKey(key)
	return btcutil.NewAddressPubKeyHash(btcutil.Hash160(btcPK.SerializeCompressed()), chain.Params())
}
