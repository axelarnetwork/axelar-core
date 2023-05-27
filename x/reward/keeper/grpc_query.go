package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils/slices"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the nexus module
type Querier struct {
	keeper Keeper
	minter types.Minter
	nexus  types.Nexus
}

// NewGRPCQuerier creates a new nexus Querier
func NewGRPCQuerier(k Keeper, m types.Minter, n types.Nexus) Querier {
	return Querier{
		keeper: k,
		minter: m,
		nexus:  n,
	}
}

// InflationRate returns the Axelar network inflation
func (q Querier) InflationRate(c context.Context, req *types.InflationRateRequest) (*types.InflationRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := q.keeper.GetParams(ctx)

	baseInflation := q.minter.GetMinter(ctx).Inflation
	keyManagementInflation := params.KeyMgmtRelativeInflationRate.Mul(baseInflation)

	chains := slices.Filter(q.nexus.GetChains(ctx), func(chain nexus.Chain) bool {
		if !q.nexus.IsChainActivated(ctx, chain) {
			return false
		}

		maintainers := q.nexus.GetChainMaintainers(ctx, chain)
		return len(maintainers) != 0
	})
	chainMaintainerInflation := params.ExternalChainVotingInflationRate.MulInt64(int64(len(chains)))

	inflation := baseInflation.Add(keyManagementInflation).Add(chainMaintainerInflation)

	return &types.InflationRateResponse{
		InflationRate: inflation,
	}, nil
}
