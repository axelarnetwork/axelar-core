package rpc

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	evmClient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . Client MoonbeamClient

const (
	codeMethodNotFound = -32601
	codeServerError    = -32000
	ganacheChainID     = int64(1337)
)

// Client provides calls to EVM JSON-RPC endpoints
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Close()
}

// MoonbeamClient provides calls to Moonbeam JSON-RPC endpoints
type MoonbeamClient interface {
	Client

	ChainGetFinalizedHead(ctx context.Context) (common.Hash, error)
	ChainGetHeader(ctx context.Context, hash common.Hash) (*MoonbeamHeader, error)
}

// MoonbeamClientImpl implements MoonbeamClient
type MoonbeamClientImpl struct {
	*evmClient.Client
	url string
}

// ChainGetFinalizedHead returns the hash of the latest finalized block
func (c MoonbeamClientImpl) ChainGetFinalizedHead(ctx context.Context) (common.Hash, error) {
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
func (c MoonbeamClientImpl) ChainGetHeader(ctx context.Context, hash common.Hash) (*MoonbeamHeader, error) {
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

// NewClient returns an EVM rpc client
func NewClient(url string) (Client, error) {
	evmClient, err := evmClient.Dial(url)
	if err != nil {
		return nil, err
	}

	moonbeamClient := MoonbeamClientImpl{
		Client: evmClient,
		url:    url,
	}

	chainID, err := evmClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	_, err = moonbeamClient.ChainGetFinalizedHead(context.Background())
	switch err := err.(type) {
	case nil:
		return moonbeamClient, nil
	case rpc.HTTPError:
		var jsonrpcMsg jsonrpcMessage
		if json.Unmarshal(err.Body, &jsonrpcMsg) != nil {
			return nil, err
		}

		if jsonrpcMsg.Error != nil && jsonrpcMsg.Error.Code == codeMethodNotFound {
			return evmClient, nil
		}

		return nil, err
	case rpc.Error:
		switch {
		case err.ErrorCode() <= codeServerError && chainID.Int64() == ganacheChainID:
			fallthrough
		case err.ErrorCode() == codeMethodNotFound:
			return evmClient, nil
		default:
			return nil, err
		}
	default:
		return nil, err
	}
}
