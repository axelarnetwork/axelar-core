package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

const (
	ModeSpecificAddress Mode = iota
	ModeCurrentMasterKey
	ModeNextMasterKey
	ModeSpecificKey
)

type Mode int

type UTXO struct {
	Hash      *chainhash.Hash
	VoutIdx   uint32
	Amount    btcutil.Amount
	Recipient BtcAddress
}

func (u UTXO) Validate() error {
	if u.Hash == nil {
		return fmt.Errorf("missing hash")
	}
	if u.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if err := u.Recipient.Validate(); err != nil {
		return err
	}
	return nil
}

func (u UTXO) Equals(other UTXO) bool {
	return u.Hash.IsEqual(other.Hash) &&
		u.VoutIdx == other.VoutIdx &&
		u.Amount == other.Amount &&
		u.Recipient == other.Recipient
}

// This type provides additional functionality based on the bitcoin chain name
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
