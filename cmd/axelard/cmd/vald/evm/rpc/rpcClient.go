package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	evmClient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . Client

const codeMethodNotFound = -32601

// Client provides calls to an EVM RPC endpoint
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Close()

	ChainGetFinalizedHead(ctx context.Context) (common.Hash, error)
	ChainGetHeader(ctx context.Context, hash common.Hash) (*MoonbeamHeader, error)
	IsMoonbeam() bool
}

// ClientImpl implements Client
type ClientImpl struct {
	*evmClient.Client
	url        string
	isMoonbeam bool
}

// ChainGetFinalizedHead returns the hash of the latest finalized block
func (c ClientImpl) ChainGetFinalizedHead(ctx context.Context) (common.Hash, error) {
	rpc, err := rpc.DialContext(ctx, c.url)
	if err != nil {
		return common.Hash{}, err
	}

	var result common.Hash
	if err := rpc.CallContext(ctx, &result, "chain_getFinalizedHead"); err != nil {
		return common.Hash{}, err
	}

	return result, nil
}

// ChainGetHeader returns the moonbeam block header of the given hash
func (c ClientImpl) ChainGetHeader(ctx context.Context, hash common.Hash) (*MoonbeamHeader, error) {
	rpc, err := rpc.DialContext(ctx, c.url)
	if err != nil {
		return nil, err
	}

	var result MoonbeamHeader
	if err := rpc.CallContext(ctx, &result, "chain_getHeader", hash); err != nil {
		return nil, err
	}

	return &result, nil
}

// IsMoonbeam returns true if the rpc client is connected to a moonbeam node; false otherwise
func (c ClientImpl) IsMoonbeam() bool {
	return c.isMoonbeam
}

// NewClient returns an EVM rpc client
func NewClient(url string) (Client, error) {
	c, err := evmClient.Dial(url)
	if err != nil {
		return nil, err
	}

	client := ClientImpl{
		Client: c,
		url:    url,
	}

	_, err = client.ChainGetFinalizedHead(context.Background())
	switch err := err.(type) {
	case nil:
		client.isMoonbeam = true

		return client, nil
	case rpc.Error:
		if err.ErrorCode() == codeMethodNotFound {
			client.isMoonbeam = false

			return client, nil
		}

		return nil, err
	default:
		return nil, err
	}
}
