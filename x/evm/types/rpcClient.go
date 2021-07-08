package types

import (
	"context"
	"fmt"
	"math/big"

	evm "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	evmClient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . RPCClient

// RPCClient provides calls to an EVM RPC endpoint
type RPCClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*evmTypes.Receipt, error)
	SendTransaction(ctx context.Context, tx *evmTypes.Transaction) error
	SendAndSignTransaction(ctx context.Context, msg evm.CallMsg) (string, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg evm.CallMsg) (uint64, error)
}

// RPCClientImpl implements RPCClient
type RPCClientImpl struct {
	*evmClient.Client
	rpc *rpc.Client
}

// NewRPCClient returns an EVM rpc client
func NewRPCClient(url string) (*RPCClientImpl, error) {
	client, err := evmClient.Dial(url)
	if err != nil {
		return nil, err
	}

	rpcClient, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}

	// try to access EVM network
	if _, err := client.ChainID(context.Background()); err != nil {
		return nil, err
	}

	return &RPCClientImpl{Client: client, rpc: rpcClient}, nil
}

// SendAndSignTransaction sends an unsigned transaction to an EVM node which tries to sign and submit it
func (evmRPCClient *RPCClientImpl) SendAndSignTransaction(ctx context.Context, msg evm.CallMsg) (string, error) {
	var txHash hexutil.Bytes

	err := evmRPCClient.rpc.CallContext(ctx, &txHash, "eth_sendTransaction", toCallArg(msg))
	if err != nil {
		return "", err
	}

	return txHash.String(), nil
}

/* Copied from https://github.com/ethereum/go-ethereum/blob/053ed9cc847647a9b3ef707d0efe7104c4ab2a4c/ethclient/ethclient.go#L531 */
func toCallArg(msg evm.CallMsg) interface{} {
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
func (d dummyClient) TransactionReceipt(context.Context, common.Hash) (*evmTypes.Receipt, error) {
	return nil, fmt.Errorf("no response")
}

// SendTransaction implements RPCClient
func (d dummyClient) SendTransaction(context.Context, *evmTypes.Transaction) error {
	return fmt.Errorf("no response")
}

// SendAndSignTransaction implements RPCClient
func (d dummyClient) SendAndSignTransaction(context.Context, evm.CallMsg) (string, error) {
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
	return big.NewInt(1), nil
}

// EstimateGas implements RPCClient
func (d dummyClient) EstimateGas(context.Context, evm.CallMsg) (uint64, error) {
	return 0, fmt.Errorf("no response")
}
