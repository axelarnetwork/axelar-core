package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	evmClient "github.com/ethereum/go-ethereum/ethclient"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . Client

// Client provides calls to an EVM RPC endpoint
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

// ClientImpl implements Client
type ClientImpl struct {
	*evmClient.Client
}

// NewClient returns an EVM rpc client
func NewClient(url string) (*evmClient.Client, error) {
	client, err := evmClient.Dial(url)
	if err != nil {
		return nil, err
	}

	// try to access network
	if _, err := client.ChainID(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}
