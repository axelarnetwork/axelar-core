package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// Ethereum2Client is a JSON-RPC client of Ethereum 2.0
type Ethereum2Client struct {
	*EthereumClient
}

// NewEthereum2Client is the constructor
func NewEthereum2Client(ethereumClient *EthereumClient) (*Ethereum2Client, error) {
	client := &Ethereum2Client{EthereumClient: ethereumClient}
	if _, err := client.latestFinalizedBlockNumber(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

// IsFinalized determines whether or not the given transaction receipt is finalized on the chain
func (c *Ethereum2Client) IsFinalized(ctx context.Context, _ uint64, txReceipt *types.Receipt) (bool, error) {
	latestFinalizedBlockNumber, err := c.latestFinalizedBlockNumber(ctx)
	if err != nil {
		return false, err
	}

	return latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) >= 0, nil
}

func (c *Ethereum2Client) latestFinalizedBlockNumber(ctx context.Context) (*big.Int, error) {
	var head *types.Header
	err := c.rpc.CallContext(ctx, &head, "eth_getBlockByNumber", "finalized", false)
	if err != nil {
		return nil, err
	}
	if head == nil {
		return nil, ethereum.NotFound
	}

	return head.Number, nil
}
