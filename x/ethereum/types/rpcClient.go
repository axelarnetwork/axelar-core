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

type rpc struct {
	*ethclient.Client
	networkID *big.Int
}

func NewRPCClient(url string) (*rpc, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	r := &rpc{Client: client}

	// cache the network id
	if _, err := r.NetworkID(context.Background()); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *rpc) NetworkID(ctx context.Context) (*big.Int, error) {
	if r.networkID == nil {
		networkID, err := r.Client.NetworkID(ctx)
		if err != nil {
			return nil, err
		}
		r.networkID = networkID
	}
	return r.networkID, nil
}
