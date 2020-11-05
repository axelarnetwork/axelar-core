package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Bridge struct {
	rpc                        *rpcclient.Client
	expectedConfirmationHeight int64
}

func NewBridge(rpc *rpcclient.Client, expectedConfirmationHeight int64) Bridge {
	return Bridge{rpc: rpc, expectedConfirmationHeight: expectedConfirmationHeight}
}

func (b Bridge) TrackAddress(address string) error {
	return b.rpc.ImportAddress(address)
}

func (b Bridge) VerifyTx(txHash *chainhash.Hash, expectedAmount btcutil.Amount) error {
	btcTxResult, err := b.rpc.GetTransaction(txHash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	actualAmount, err := btcutil.NewAmount(btcTxResult.Amount)
	if err != nil {
		return sdkerrors.Wrap(err, "could not parse transaction amount of the Bitcoin response")
	}

	isEqual := btcTxResult.TxID == txHash.String() &&
		expectedAmount == actualAmount &&
		btcTxResult.Confirmations >= b.expectedConfirmationHeight
	if !isEqual {
		return fmt.Errorf(
			"transaction on Bitcoin differs from expected transaction: {txID: %s, amount: %v, destination: %s}",
			btcTxResult.TxID, btcTxResult.Amount, btcTxResult.Details[0].Address,
		)
	}
	return nil
}

func (b Bridge) Send(tx *wire.MsgTx) error {
	_, err := b.rpc.SendRawTransaction(tx, false)
	return err
}
