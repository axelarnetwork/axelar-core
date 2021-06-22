package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the broadcast MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServiceServer {
	return msgServer{Keeper: keeper}
}

func (s msgServer) RegisterProxy(c context.Context, req *types.RegisterProxyRequest) (*types.RegisterProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := s.Keeper.RegisterProxy(ctx, req.PrincipalAddr, req.ProxyAddr); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBroadcast, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeRegisterProxy),
			sdk.NewAttribute(sdk.AttributeKeySender, req.PrincipalAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, req.ProxyAddr.String()),
		),
	)
	return &types.RegisterProxyResponse{}, nil
}

func (s msgServer) DeregisterProxy(c context.Context, req *types.DeregisterProxyRequest) (*types.DeregisterProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	proxy := s.Keeper.GetProxy(ctx, req.PrincipalAddr)

	if err := s.Keeper.DeregisterProxy(ctx, req.PrincipalAddr); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBroadcast, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeDeregisterProxy),
			sdk.NewAttribute(sdk.AttributeKeySender, req.PrincipalAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, proxy.String()),
		),
	)

	s.Keeper.Logger(ctx).Info("validator %s has de-registered proxy %s", req.PrincipalAddr, proxy)
	return &types.DeregisterProxyResponse{}, nil
}
