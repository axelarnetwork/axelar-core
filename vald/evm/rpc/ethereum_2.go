package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

type ethereum2Client struct {
	*ethereumClient
}

func newEthereum2Client(ethereumClient *ethereumClient) (*ethereum2Client, error) {
	client := &ethereum2Client{ethereumClient: ethereumClient}
	if _, err := client.LatestFinalizedBlockNumber(context.Background(), 0); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ethereum2Client) LatestFinalizedBlockNumber(ctx context.Context, _ uint64) (*big.Int, error) {
	var head *types.Header
	if err := c.rpc.CallContext(ctx, &head, "eth_getBlockByNumber", "finalized", false); err != nil {
		return nil, err
	}

	return head.Number, nil
}
