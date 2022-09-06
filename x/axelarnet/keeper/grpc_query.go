package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc querier
type Querier struct {
	nexus  types.Nexus
	keeper types.BaseKeeper
}

// NewGRPCQuerier returns a new Querier
func NewGRPCQuerier(k types.BaseKeeper, n types.Nexus) Querier {
	return Querier{
		keeper: k,
		nexus:  n,
	}
}

// PendingIBCTransferCount returns the number of pending IBC transfers per Cosmos chain, upto the transfer limit
func (q Querier) PendingIBCTransferCount(c context.Context, _ *types.PendingIBCTransferCountRequest) (*types.PendingIBCTransferCountResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chains := q.keeper.GetCosmosChains(ctx)
	counts := make(map[string]uint32, len(chains))
	for _, c := range chains {
		chain, ok := q.nexus.GetChain(ctx, c)
		if !ok {
			return nil, fmt.Errorf("cosmos chain %s not found in the %s module", c, nexusTypes.ModuleName)
		}
		pageRequest := &query.PageRequest{
			Key:        nil,
			Offset:     0,
			Limit:      1, // 1 minimizes the number of values processed, and is independent of counting the total matches
			CountTotal: true,
			Reverse:    false,
		}
		_, resp, err := q.nexus.GetTransfersForChainPaginated(ctx, chain, nexus.Pending, pageRequest)
		if err != nil {
			return nil, err
		}
		counts[c.String()] = uint32(resp.Total) // assert: there should never be more than 4294967295 transfers in the queue
	}

	return &types.PendingIBCTransferCountResponse{TransfersByChain: counts}, nil
}
