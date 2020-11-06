package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

type UTXO struct {
	Hash    *chainhash.Hash
	VoutIdx uint32
	Amount  btcutil.Amount
	Address btcutil.Address
}

func (u UTXO) PkScript() []byte {
	if script, err := txscript.PayToAddrScript(u.Address); err != nil {
		return nil
	} else {
		return script
	}
}

func (u UTXO) IsInvalid() bool {
	return u.Hash == nil || u.Amount < 0 || u.Address == nil
}

func (u UTXO) Equals(other UTXO) bool {
	return u.Hash.IsEqual(other.Hash) &&
		u.VoutIdx == other.VoutIdx &&
		u.Amount == other.Amount &&
		u.Address.String() == other.Address.String()
}
