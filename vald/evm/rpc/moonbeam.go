package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type moonbeamClient struct {
	*ethereumClient
}

func newMoonbeamClient(ethereumClient *ethereumClient) (*moonbeamClient, error) {
	client := &moonbeamClient{ethereumClient: ethereumClient}
	if _, err := client.latestFinalizedBlockNumber(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *moonbeamClient) IsFinalized(ctx context.Context, _ uint64, txReceipt *types.Receipt) (bool, error) {
	latestFinalizedBlockNumber, err := c.latestFinalizedBlockNumber(ctx)
	if err != nil {
		return false, err
	}

	return latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) >= 0, nil
}

func (c *moonbeamClient) latestFinalizedBlockNumber(ctx context.Context) (*big.Int, error) {
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
