package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type moonbeamClient struct {
	*ethereumClient
}

func newMoonbeamClient(ethereumClient *ethereumClient) (*moonbeamClient, error) {
	client := &moonbeamClient{ethereumClient: ethereumClient}
	if _, err := client.LatestFinalizedBlockNumber(context.Background(), 0); err != nil {
		return nil, err
	}

	return client, nil
}

func (c moonbeamClient) LatestFinalizedBlockNumber(ctx context.Context, _ uint64) (*big.Int, error) {
	var blockHash common.Hash
	if err := c.rpc.CallContext(ctx, &blockHash, "chain_getFinalizedHead"); err != nil {
		return nil, err
	}

	var moonbeamHeader MoonbeamHeader
	if err := c.rpc.CallContext(ctx, &moonbeamHeader, "chain_getHeader", blockHash); err != nil {
		return nil, err
	}

	return moonbeamHeader.Number.ToInt(), nil
}
