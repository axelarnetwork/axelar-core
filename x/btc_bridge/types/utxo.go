package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

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
