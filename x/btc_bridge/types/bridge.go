package types

import (
	"fmt"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Bridge struct {
	rpc                        *rpcclient.Client
	expectedConfirmationHeight uint64
}

func NewBridge(rpc *rpcclient.Client, expectedConfirmationHeight uint64) Bridge {
	return Bridge{rpc: rpc, expectedConfirmationHeight: expectedConfirmationHeight}
}

func (b Bridge) TrackAddress(address string) error {
	return b.rpc.ImportAddress(address)
}

func (b Bridge) VerifyTx(utxo UTXO) error {
	actualTx, err := b.rpc.GetRawTransactionVerbose(utxo.Hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	if utxo.VoutIdx >= uint32(len(actualTx.Vout)) {
		return fmt.Errorf("vout index out of range")
	}

	vout := actualTx.Vout[utxo.VoutIdx]

	if len(vout.ScriptPubKey.Addresses) > 1 {
		return fmt.Errorf("deposit must be only spendable by a single address")
	}
	if vout.ScriptPubKey.Addresses[0] != utxo.Address.String() {
		return fmt.Errorf("expected destination address does not match actual destination address")
	}

	actualAmount, err := btcutil.NewAmount(vout.Value)
	if err != nil {
		return sdkerrors.Wrap(err, "could not parse transaction amount of the Bitcoin response")
	}
	if utxo.Amount != actualAmount {
		return fmt.Errorf("expected amount does not match actual amount")
	}

	if actualTx.Confirmations < b.expectedConfirmationHeight {
		return fmt.Errorf("not enough confirmations yet")
	}

	return nil
}

func (b Bridge) Send(tx *wire.MsgTx) error {
	_, err := b.rpc.SendRawTransaction(tx, false)
	return err
}
