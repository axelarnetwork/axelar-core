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
	validator := req.Validator

	chains := slices.Filter(q.nexus.GetChains(ctx), func(chain nexus.Chain) bool {
		if !q.nexus.IsChainActivated(ctx, chain) {
			return false
		}

		maintainers := q.nexus.GetChainMaintainers(ctx, chain)
		if len(validator) == 0 {
			return len(maintainers) > 0
		}

		return slices.Any(maintainers, func(maintainer sdk.ValAddress) bool { return validator.Equals(maintainer) })
	})
	chainMaintainerInflation := params.ExternalChainVotingInflationRate.MulInt64(int64(len(chains)))

	inflation := baseInflation.Add(keyManagementInflation).Add(chainMaintainerInflation)

	return &types.InflationRateResponse{
		InflationRate: inflation,
	}, nil
}

// Params returns the reward module params
func (q Querier) Params(c context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := q.keeper.GetParams(ctx)

	return &types.ParamsResponse{
		Params: params,
	}, nil
}
