package rpc

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate moq -out ./mock/ethereum.go -pkg mock . EthereumJSONRPCClient JSONRPCClient

// EthereumJSONRPCClient represents the functionality of github.com/ethereum/go-ethereum/ethclient.Client
type EthereumJSONRPCClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	Close()
}

// JSONRPCClient represents the functionality of github.com/ethereum/go-ethereum/rpc.Client
type JSONRPCClient interface {
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
}

// EthereumClient is a JSON-RPC client of any Ethereum-compact chains
type EthereumClient struct {
	EthereumJSONRPCClient
	rpc JSONRPCClient
}

// NewEthereumClient is the constructor
func NewEthereumClient(ethClient EthereumJSONRPCClient, rpc JSONRPCClient) (*EthereumClient, error) {
	client := &EthereumClient{
		EthereumJSONRPCClient: ethClient,
		rpc:                   rpc,
	}
	// validate that the given url implements standard ethereum JSON-RPC
	if _, err := client.BlockNumber(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

// HeaderByNumber returns the block header for the given block number
func (c *EthereumClient) HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error) {
	var head *Header
	err := c.rpc.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}

	return head, err
}

// IsFinalized determines whether or not the given transaction receipt is finalized on the chain
func (c *EthereumClient) IsFinalized(ctx context.Context, conf uint64, txReceipt *types.Receipt) (bool, error) {
	blockNumber, err := c.BlockNumber(ctx)
	if err != nil {
		return false, err
	}

	latestFinalizedBlockNumber := sdk.NewIntFromUint64(blockNumber).SubRaw(int64(conf)).AddRaw(1).BigInt()

	return latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) >= 0, nil
}

// copied from https://github.com/ethereum/go-ethereum/blob/69568c554880b3567bace64f8848ff1be27d084d/ethclient/ethclient.go#L565
func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}

	return hexutil.EncodeBig(number)
}
