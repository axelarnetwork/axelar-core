package rpc

import (
	"context"
	"math/big"
)

// ConfirmationIfAheadClient is a JSON-RPC client that uses the RPC finalized tag, unless the confirmation height is ahead of it
type ConfirmationIfAheadClient struct {
	*Ethereum2Client
}

// NewConfirmationIfAheadClient is the constructor
func NewConfirmationIfAheadClient(ethereum2Client *Ethereum2Client) *ConfirmationIfAheadClient {
	return &ConfirmationIfAheadClient{Ethereum2Client: ethereum2Client}
}

// LatestFinalizedBlockNumber returns the higher of the RPC finalized block and the block at the given confirmation height.
// If only one of the two can be determined, that one is returned.
func (c *ConfirmationIfAheadClient) LatestFinalizedBlockNumber(ctx context.Context, confirmations uint64) (*big.Int, error) {
	finalized, finalizedErr := c.Ethereum2Client.LatestFinalizedBlockNumber(ctx, confirmations)
	confirmed, confirmedErr := c.EthereumClient.LatestFinalizedBlockNumber(ctx, confirmations)

	switch {
	case finalizedErr != nil && confirmedErr != nil:
		return nil, finalizedErr
	case finalizedErr != nil:
		return confirmed, nil
	case confirmedErr != nil:
		return finalized, nil
	case finalized.Cmp(confirmed) >= 0:
		return finalized, nil
	default:
		return confirmed, nil
	}
}
