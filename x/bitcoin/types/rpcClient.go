package types

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . RPCClient

// RPCClient provides calls to an Bitcoin RPC endpoint
type RPCClient interface {
	ListUnspent() ([]btcjson.ListUnspentResult, error)
	EstimateSmartFee(confTarget int64, mode *btcjson.EstimateSmartFeeMode) (*btcjson.EstimateSmartFeeResult, error)
	SignRawTransactionWithWallet(tx *wire.MsgTx) (*wire.MsgTx, bool, error)
}
