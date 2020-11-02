package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
)

const (
	Satoshi int64 = 1
	Bitcoin       = 100_000_000 * Satoshi
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

func (b Bridge) VeriyfyTx(tx exported.ExternalTx) error {
	hash, err := chainhash.NewHashFromStr(tx.TxID)
	if err != nil {
		return sdkerrors.Wrap(err, "could not transform Bitcoin transaction ID to hash")
	}

	btcTxResult, err := b.rpc.GetTransaction(hash)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction")
	}

	actualAmount, err := btcutil.NewAmount(btcTxResult.Amount)
	if err != nil {
		return sdkerrors.Wrap(err, "could not parse transaction amount of the Bitcoin response")
	}

	expectedAmount := tx.Amount.Amount
	isEqual := btcTxResult.TxID == tx.TxID &&
		amountEquals(expectedAmount, actualAmount) &&
		btcTxResult.Confirmations >= b.expectedConfirmationHeight
	if !isEqual {
		return fmt.Errorf(
			"transaction on Bitcoin differs from expected transaction: {txID: %s, amount: %v, destination: %s}",
			btcTxResult.TxID, btcTxResult.Amount, btcTxResult.Details[0].Address,
		)
	}
	return nil
}

func amountEquals(expectedAmount sdk.Dec, actualAmount btcutil.Amount) bool {
	return (expectedAmount.IsInteger() && satoshiEquals(expectedAmount, actualAmount)) ||
		btcEquals(expectedAmount, actualAmount)
}

func satoshiEquals(satoshiAmount sdk.Dec, verifiedAmount btcutil.Amount) bool {
	return satoshiAmount.IsInt64() && btcutil.Amount(satoshiAmount.Int64()) == verifiedAmount
}

func btcEquals(btcAmount sdk.Dec, verifiedAmount btcutil.Amount) bool {
	return btcutil.Amount(btcAmount.MulInt64(Bitcoin).RoundInt64()) == verifiedAmount
}
