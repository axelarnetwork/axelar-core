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
	if _, err := client.LatestFinalizedBlockNumber(context.Background(), 0); err != nil {
		return nil, err
	}

	return client, nil
}

// LatestFinalizedBlockNumber returns the latest finalized block number
func (c *Ethereum2Client) LatestFinalizedBlockNumber(ctx context.Context, _ uint64) (*big.Int, error) {
	var head *types.Header
	err := c.rpc.CallContext(ctx, &head, "eth_getBlockByNumber", "finalized", false)
	if err != nil {
		return nil, err
	}
	if head == nil || head.Number == nil {
		return nil, ethereum.NotFound
	}

	return head.Number, nil
}
