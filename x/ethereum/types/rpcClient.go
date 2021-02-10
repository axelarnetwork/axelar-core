package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . RPCClient

// RPCClient provides calls to an Ethereum RPC endpoint
type RPCClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	SendAndSignTransaction(ctx context.Context, msg ethereum.CallMsg) (string, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

// RPCClientImpl implements RPCClient
type RPCClientImpl struct {
	*ethclient.Client
	rpc *rpc.Client
}

// NewRPCClient returns an Ethereum rpc client
func NewRPCClient(url string) (*RPCClientImpl, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	rpcClient, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}

	// try to access ethereum network
	if _, err := client.ChainID(context.Background()); err != nil {
		return nil, err
	}

	return &RPCClientImpl{Client: client, rpc: rpcClient}, nil
}

// SendAndSignTransaction sends an unsigned transaction to an Ethereum node which tries to sign and submit it
func (ethRPCClient *RPCClientImpl) SendAndSignTransaction(ctx context.Context, msg ethereum.CallMsg) (string, error) {
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

type dummyClient struct{}

// NewDummyRPC returns a placeholder for an rpc client. It does not make any rpc calls
func NewDummyRPC() RPCClient {
	return dummyClient{}
}

// BlockNumber implements RPCClient
func (d dummyClient) BlockNumber(context.Context) (uint64, error) {
	return 0, fmt.Errorf("no response")
}

// TransactionReceipt implements RPCClient
func (d dummyClient) TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return nil, fmt.Errorf("no response")
}

// SendTransaction implements RPCClient
func (d dummyClient) SendTransaction(context.Context, *types.Transaction) error {
	return fmt.Errorf("no response")
}

// SendAndSignTransaction implements RPCClient
func (d dummyClient) SendAndSignTransaction(context.Context, ethereum.CallMsg) (string, error) {
	return "", fmt.Errorf("no response")
}

// PendingNonceAt implements RPCClient
func (d dummyClient) PendingNonceAt(context.Context, common.Address) (uint64, error) {
	return 0, fmt.Errorf("no response")
}

// SuggestGasPrice implements RPCClient
func (d dummyClient) SuggestGasPrice(context.Context) (*big.Int, error) {
	return nil, fmt.Errorf("no response")
}

// ChainID implements RPCClient
func (d dummyClient) ChainID(context.Context) (*big.Int, error) {
	return DefaultParams().Network.Params().ChainID, nil
}

// EstimateGas implements RPCClient
func (d dummyClient) EstimateGas(context.Context, ethereum.CallMsg) (uint64, error) {
	return 0, fmt.Errorf("no response")
}

// CallContext implements RPCClient
func (d dummyClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return nil, fmt.Errorf("no response")
}
