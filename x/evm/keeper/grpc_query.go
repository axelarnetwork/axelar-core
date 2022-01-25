package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServiceServer = BaseKeeper{}

// BurnerInfo implements the burner info grpc query
func (k BaseKeeper) BurnerInfo(c context.Context, req *types.BurnerInfoRequest) (*types.BurnerInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	iter := k.getBaseStore(ctx).Iterator(subspacePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		ck := k.ForChain(string(iter.Value()))
		burnerInfo := ck.GetBurnerInfo(ctx, req.Address)
		if burnerInfo != nil {
			return &types.BurnerInfoResponse{Chain: ck.GetParams(ctx).Chain, BurnerInfo: burnerInfo}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "unknown address")
}
