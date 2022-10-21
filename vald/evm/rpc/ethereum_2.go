package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

type ethereum2Client struct {
	*ethereumClient
}

func newEthereum2Client(ethereumClient *ethereumClient) (*ethereum2Client, error) {
	client := &ethereum2Client{ethereumClient: ethereumClient}
	if _, err := client.latestFinalizedBlockNumber(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ethereum2Client) IsFinalized(ctx context.Context, _ uint64, txReceipt *types.Receipt) (bool, error) {
	latestFinalizedBlockNumber, err := c.latestFinalizedBlockNumber(ctx)
	if err != nil {
		return false, err
	}

	return latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) >= 0, nil
}

func (c *ethereum2Client) latestFinalizedBlockNumber(ctx context.Context) (*big.Int, error) {
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
