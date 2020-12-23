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

type ExternalChainAddress struct {
	Chain   string
	Address string
}

func (addr ExternalChainAddress) IsInvalid() bool {
	return addr.Chain == "" || addr.Address == ""
}

func (addr ExternalChainAddress) String() string {
	return fmt.Sprintf("chain: %s, address: %s", addr.Chain, addr.Address)
}

type UTXO struct {
	Hash    *chainhash.Hash
	VoutIdx uint32
	Amount  btcutil.Amount
	Address BtcAddress
}

func (u UTXO) Validate() error {
	if u.Hash == nil {
		return fmt.Errorf("missing hash")
	}
	if u.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if err := u.Address.Validate(); err != nil {
		return err
	}
	return nil
}

func (u UTXO) Equals(other UTXO) bool {
	return u.Hash.IsEqual(other.Hash) &&
		u.VoutIdx == other.VoutIdx &&
		u.Amount == other.Amount &&
		u.Address == other.Address
}

// This type provides additional functionality based on the bitcoin chain name
type Chain string

// Validate checks if the object is a valid chain
func (c Chain) Validate() error {
	switch string(c) {
	case chaincfg.MainNetParams.Name, chaincfg.TestNet3Params.Name:
		return nil
	default:
		return fmt.Errorf("chain could not be parsed, choose %s or %s",
			chaincfg.MainNetParams.Name,
			chaincfg.TestNet3Params.Name,
		)
	}
}

// Params returns the configuration parameters associated with the chain
func (c Chain) Params() *chaincfg.Params {
	switch string(c) {
	case chaincfg.MainNetParams.Name:
		return &chaincfg.MainNetParams
	case chaincfg.TestNet3Params.Name:
		return &chaincfg.TestNet3Params
	default:
		return nil
	}
}
