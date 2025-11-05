package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

var _ types.QueryServer = Keeper{}

// GovernanceKey returns the multisig governance key
func (k Keeper) GovernanceKey(c context.Context, req *types.QueryGovernanceKeyRequest) (*types.QueryGovernanceKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	governanceKey, ok := k.GetGovernanceKey(ctx)
	if !ok {
		return nil, fmt.Errorf("governance key not set")
	}

	return &types.QueryGovernanceKeyResponse{
		GovernanceKey: governanceKey,
	}, nil
}

// Params returns the reward module params
func (q Keeper) Params(c context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := q.GetParams(ctx)

	return &types.ParamsResponse{
		Params: params,
	}, nil
}
