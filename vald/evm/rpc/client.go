package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate moq -out ./mock/client.go -pkg mock . Client

// Client provides calls to EVM JSON-RPC endpoints
type Client interface {
	// TransactionReceipt returns the transaction receipt for the given transaction hash
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	// HeaderByNumber returns the block header for the given block number
	HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error)
	// LatestFinalizedBlockNumber returns the latest finalized block number with the given number of confirmations if no true finality is supported by the chain
	LatestFinalizedBlockNumber(ctx context.Context, conf uint64) (*big.Int, error)
	// Close closes the client connection
	Close()
}

// NewClient returns an EVM rpc client
func NewClient(url string) (Client, error) {
	ethereumClient, err := newEthereumClient(url)
	if err != nil {
		return nil, err
	}

	if moonbeamClient, err := newMoonbeamClient(ethereumClient); err == nil {
		return moonbeamClient, nil
	}

	if ethereum2Client, err := newEthereum2Client(ethereumClient); err == nil {
		return ethereum2Client, nil
	}

	return ethereumClient, nil
}
