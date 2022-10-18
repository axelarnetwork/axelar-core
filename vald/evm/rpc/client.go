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
	// IsFinalized determines whether or not the given transaction receipt is finalized on the chain
	IsFinalized(ctx context.Context, conf uint64, txReceipt *types.Receipt) (bool, error)
	// Close closes the client connection
	Close()
}

// NewClient returns an EVM JSON-RPC client
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

// NewL2Client returns a L2 EVM JSON-RPC client
func NewL2Client(url string, l1Client Client) (Client, error) {
	ethereumClient, err := newEthereumClient(url)
	if err != nil {
		return nil, err
	}

	return newArbitrumClient(ethereumClient, l1Client)
}
