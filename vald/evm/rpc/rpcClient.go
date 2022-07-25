package rpc

import (
	"context"
	"math/big"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	evmClient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . Client MoonbeamClient

// Client provides calls to EVM JSON-RPC endpoints
type Client interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	Close()
}

// Eth2Client provides calls to Ethereum JSON-RPC endpoints post the merge
type Eth2Client interface {
	Client

	FinalizedHeader(ctx context.Context) (*types.Header, error)
}

// Eth2ClientImpl implements Eth2Client
type Eth2ClientImpl struct {
	*evmClient.Client
	url string
}

// FinalizedHeader returns the header of the most recent finalized block
func (c Eth2ClientImpl) FinalizedHeader(ctx context.Context) (*types.Header, error) {
	rpc, err := rpc.DialContext(ctx, c.url)
	if err != nil {
		return nil, err
	}

	var head *types.Header
	if err := rpc.CallContext(ctx, &head, "eth_getBlockByNumber", "finalized", false); err != nil {
		return nil, err
	}

	return head, nil
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

	// validate that the given url implements standard ethereum JSON-RPC
	if _, err := evmClient.BlockNumber(context.Background()); err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot query latest block number with the given EVM JSON-RPC url %s", url)
	}

	moonbeamClient := MoonbeamClientImpl{
		Client: evmClient,
		url:    url,
	}
	if _, err := moonbeamClient.ChainGetFinalizedHead(context.Background()); err == nil {
		return moonbeamClient, nil
	}

	eth2Client := Eth2ClientImpl{
		Client: evmClient,
		url:    url,
	}

	// TODO: not return a eth2 client after the merge is settled on ethereum mainnet
	return eth2Client, nil
}
