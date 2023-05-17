package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// MoonbeamClient is a JSON-RPC client of Moonbeam
type MoonbeamClient struct {
	*EthereumClient
}

// NewMoonbeamClient is the constructor
func NewMoonbeamClient(ethereumClient *EthereumClient) (*MoonbeamClient, error) {
	client := &MoonbeamClient{EthereumClient: ethereumClient}
	if _, err := client.LatestFinalizedBlockNumber(context.Background(), 0); err != nil {
		return nil, err
	}

	return client, nil
}

// LatestFinalizedBlockNumber returns the latest finalized block number
func (c *MoonbeamClient) LatestFinalizedBlockNumber(ctx context.Context, _ uint64) (*big.Int, error) {
	var blockHash common.Hash
	if err := c.rpc.CallContext(ctx, &blockHash, "chain_getFinalizedHead"); err != nil {
		return nil, err
	}

	var moonbeamHeader moonbeamHeader
	if err := c.rpc.CallContext(ctx, &moonbeamHeader, "chain_getHeader", blockHash); err != nil {
		return nil, err
	}

	result := moonbeamHeader.Number.ToInt()
	if result == nil {
		return nil, ethereum.NotFound
	}

	return result, nil
}
