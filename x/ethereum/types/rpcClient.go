package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:generate moq -out ./mock/rpcClient.go -pkg mock . RPCClient

// TODO: fetch this parameters from config file, and check how to connect to actual node
const (
	myproject = "82e8e37695ed406cb9313ec09bae18e7"
	gateway   = "goerli.infura.io"
	ganache   = "http://127.0.0.1:7545"
)

func NewRPCClient(url string) (*ethclient.Client, error) {
	//return ethclient.Dial(fmt.Sprintf("https://%s/v3/%s", gateway, myproject))
	//return ethclient.Dial("http://host.docker.internal:7545")
	return ethclient.Dial(url)
}

type RPCClient interface {
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	BlockNumber(ctx context.Context) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	NetworkID(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
}
