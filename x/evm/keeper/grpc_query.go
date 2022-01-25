package keeper

import (
	"context"
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.QueryServiceServer = BaseKeeper{}

func (k BaseKeeper) BurnerInfo(c context.Context, req *types.BurnerInfoRequest) (*types.BurnerInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if !k.HasChain(ctx, req.Chain) {
		return nil, sdkerrors.Wrapf(types.ErrBurnerInfoNotFound, "unkown chain '%s'", req.Chain)
	}

	burnerInfo := k.ForChain(req.Chain).GetBurnerInfo(ctx, req.Address)
	if burnerInfo == nil {
		return nil, sdkerrors.Wrap(types.ErrBurnerInfoNotFound, fmt.Sprintf("unknown address '%s'", req.Address))
	}

	return &types.BurnerInfoResponse{BurnerInfo: burnerInfo}, nil
}
