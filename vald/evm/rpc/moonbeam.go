package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MoonbeamClient is a JSON-RPC client of Moonbeam
type MoonbeamClient struct {
	*EthereumClient
}

// NewMoonbeamClient is the constructor
func NewMoonbeamClient(ethereumClient *EthereumClient) (*MoonbeamClient, error) {
	client := &MoonbeamClient{EthereumClient: ethereumClient}
	if _, err := client.latestFinalizedBlockNumber(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

// IsFinalized determines whether or not the given transaction receipt is finalized on the chain
func (c *MoonbeamClient) IsFinalized(ctx context.Context, _ uint64, txReceipt *types.Receipt) (bool, error) {
	latestFinalizedBlockNumber, err := c.latestFinalizedBlockNumber(ctx)
	if err != nil {
		return false, err
	}

	if latestFinalizedBlockNumber == nil {
		return false, ethereum.NotFound
	}

	return latestFinalizedBlockNumber.Cmp(txReceipt.BlockNumber) >= 0, nil
}

func (c *MoonbeamClient) latestFinalizedBlockNumber(ctx context.Context) (*big.Int, error) {
	var blockHash common.Hash
	if err := c.rpc.CallContext(ctx, &blockHash, "chain_getFinalizedHead"); err != nil {
		return nil, err
	}

	var moonbeamHeader moonbeamHeader
	if err := c.rpc.CallContext(ctx, &moonbeamHeader, "chain_getHeader", blockHash); err != nil {
		return nil, err
	}

	return moonbeamHeader.Number.ToInt(), nil
}
