package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/feepolicy/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns a new msg server instance
func NewMsgServerImpl(keeper Keeper) types.MsgServiceServer {
	return msgServer{Keeper: keeper}
}

func (s msgServer) UpdateParams(c context.Context, req *types.UpdateParamsRequest) (*types.UpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := req.Params.Validate(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	s.SetParams(ctx, req.Params)
	return &types.UpdateParamsResponse{}, nil
}
