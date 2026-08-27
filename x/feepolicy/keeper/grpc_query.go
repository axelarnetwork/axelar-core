package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/feepolicy/types"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the feepolicy module
type Querier struct {
	keeper Keeper
}

// NewGRPCQuerier creates a new feepolicy Querier
func NewGRPCQuerier(k Keeper) Querier {
	return Querier{keeper: k}
}

// Params returns the policyfees module params
func (q Querier) Params(c context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := q.keeper.GetParams(ctx)

	return &types.ParamsResponse{
		Params: params,
	}, nil
}
