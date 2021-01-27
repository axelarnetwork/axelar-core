package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . RPCClient

type EthRPCClient struct {
	*ethclient.Client
	rpc *rpc.Client
}

type RPCClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	SendAndSignTransaction(ctx context.Context, msg ethereum.CallMsg) (string, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
}

func (ethRPCClient *EthRPCClient) SendAndSignTransaction(ctx context.Context, msg ethereum.CallMsg) (string, error) {
	var txHash hexutil.Bytes

	err := ethRPCClient.rpc.CallContext(ctx, &txHash, "eth_sendTransaction", toCallArg(msg))
	if err != nil {
		return "", err
	}

	return txHash.String(), nil
}

/* Copied from https://github.com/ethereum/go-ethereum/blob/053ed9cc847647a9b3ef707d0efe7104c4ab2a4c/ethclient/ethclient.go#L531 */
func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}

	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}

	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}

	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}

	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}

	return arg
}

func NewRPCClient(url string) (RPCClient, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	rpc, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}

	// try to access ethereum network
	if _, err := client.ChainID(context.Background()); err != nil {
		return nil, err
	}

	return &EthRPCClient{Client: client, rpc: rpc}, nil
}

type DummyClient struct{}

func (d DummyClient) BlockNumber(ctx context.Context) (uint64, error) {
	panic("implement me")
}

func (d DummyClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	panic("implement me")
}

func (d DummyClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	panic("implement me")
}

func (d DummyClient) SendAndSignTransaction(ctx context.Context, msg ethereum.CallMsg) (string, error) {
	panic("implement me")
}

func (d DummyClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	panic("implement me")
}

func (d DummyClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	panic("implement me")
}

func (d DummyClient) ChainID(ctx context.Context) (*big.Int, error) {
	panic("implement me")
}

func (d DummyClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	panic("implement me")
}

func NewDummyClient() (RPCClient, error) {
	return DummyClient{}, nil
}
