package types

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

type UTXO struct {
	Chain   string
	Hash    *chainhash.Hash
	VoutIdx uint32
	Amount  btcutil.Amount
	Address string
}

func (u UTXO) PkScript() []byte {
	var params *chaincfg.Params
	if u.Chain == chaincfg.MainNetParams.Name {
		params = &chaincfg.MainNetParams
	} else {
		params = &chaincfg.TestNet3Params
	}
	addr, _ := btcutil.DecodeAddress(u.Address, params)
	if script, err := txscript.PayToAddrScript(addr); err != nil {
		return nil
	} else {
		return script
	}
}

func (u UTXO) IsInvalid() bool {
	return u.Hash == nil || u.Amount < 0 || u.Address == ""
}

func (u UTXO) Equals(other UTXO) bool {
	return u.Hash.IsEqual(other.Hash) &&
		u.VoutIdx == other.VoutIdx &&
		u.Amount == other.Amount &&
		u.Address == other.Address
}
