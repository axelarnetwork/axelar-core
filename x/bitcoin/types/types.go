package types

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	ModeSpecificAddress Mode = iota
	ModeCurrentMasterKey
	ModeNextMasterKey
	ModeSpecificKey
)

type Mode int

type OutPointInfo struct {
	OutPoint      *wire.OutPoint
	Amount        btcutil.Amount
	Recipient     BtcAddress
	Confirmations uint64
}

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

func (u OutPointInfo) Equals(other OutPointInfo) bool {
	return u.OutPoint.Hash.IsEqual(&other.OutPoint.Hash) &&
		u.OutPoint.Index == other.OutPoint.Index &&
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

func ParseVoutIdx(voutIdx string) (uint32, error) {
	n, err := strconv.ParseUint(voutIdx, 10, 32)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not parse voutIdx")
	}
	return uint32(n), nil
}
