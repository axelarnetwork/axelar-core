package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModeSpecificAddress Mode = iota
	ModeCurrentMasterKey
	ModeSpecificKey
)

var (
	Mainnet  = Network(chaincfg.MainNetParams.Name)
	Testnet3 = Network(chaincfg.TestNet3Params.Name)
	Regtest  = Network(chaincfg.RegressionNetParams.Name)
)

type Mode int

// OutPointInfo describes all the necessary information to verify the outPoint of a transaction
type OutPointInfo struct {
	OutPoint      *wire.OutPoint
	Amount        btcutil.Amount
	Recipient     string
	Confirmations uint64
}

// Validate ensures that all fields are filled with sensible values
func (i OutPointInfo) Validate() error {
	if i.OutPoint == nil {
		return fmt.Errorf("missing outpoint")
	}
	if i.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if i.Recipient == "" {
		return fmt.Errorf("invalid address to track")
	}
	return nil
}

// Equals checks if two OutPointInfo objects are semantically equal
func (i OutPointInfo) Equals(other OutPointInfo) bool {
	return i.OutPoint.Hash.IsEqual(&other.OutPoint.Hash) &&
		i.OutPoint.Index == other.OutPoint.Index &&
		i.Amount == other.Amount &&
		i.Recipient == other.Recipient
}

// Network provides additional functionality based on the bitcoin network name
type Network string

// Validate checks if the object is a valid chain
func (n Network) Validate() error {
	if n.Params() == nil {
		return fmt.Errorf("network could not be parsed, choose %s, %s, or %s",
			Mainnet, Testnet3, Regtest)
	}
	return nil
}

// Params returns the configuration parameters associated with the chain
func (n Network) Params() *chaincfg.Params {
	switch n {
	case Mainnet:
		return &chaincfg.MainNetParams
	case Testnet3:
		return &chaincfg.TestNet3Params
	case Regtest:
		return &chaincfg.RegressionNetParams
	default:
		return nil
	}
}

type SendParams struct {
	SignatureID string
	TxID        string
}

type RawParams struct {
	Recipient string
	TxID      string
	Satoshi   sdk.Coin
}
