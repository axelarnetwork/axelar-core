package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/axelarnetwork/utils/monads/results"
)

//go:generate moq -out ./mock/client.go -pkg mock . Client

// Result is a custom type that allows moq to correctly generate the mock for
// results.Result with *types.Receipt.
type Result results.Result[*types.Receipt]

// Client provides calls to EVM JSON-RPC endpoints
type Client interface {
	// TransactionReceipt returns the transaction receipt for the given transaction hash
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	// TransactionReceipts returns transaction receipts for the given transaction hashes
	TransactionReceipts(ctx context.Context, txHashes []common.Hash) ([]Result, error)
	// HeaderByNumber returns the block header for the given block number
	HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error)
	// LatestFinalizedBlockNumber returns the latest finalized block number
	LatestFinalizedBlockNumber(ctx context.Context, confirmations uint64) (*big.Int, error)
	// Close closes the client connection
	Close()
}

// NewClient returns an EVM JSON-RPC client
func NewClient(url string, override FinalityOverride) (Client, error) {
	rpc, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}

	ethereumClient, err := NewEthereumClient(ethclient.NewClient(rpc), rpc)
	if err != nil {
		return nil, err
	}

	if override == Confirmation {
		return ethereumClient, nil
	}

	if ethereum2Client, err := NewEthereum2Client(ethereumClient); err == nil {
		return ethereum2Client, nil
	}

	return ethereumClient, nil
}
