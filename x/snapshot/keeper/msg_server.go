package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the snapshot MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServiceServer {
	return msgServer{Keeper: keeper}
}

func (s msgServer) RegisterProxy(c context.Context, req *types.RegisterProxyRequest) (*types.RegisterProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := s.Keeper.RegisterProxy(ctx, req.PrincipalAddr, req.ProxyAddr); err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, err.Error())
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

	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "register", "proxy"},
		float32(time.Now().Unix()),
		[]metrics.Label{
			telemetry.NewLabel("principal_address", req.PrincipalAddr.String()),
			telemetry.NewLabel("proxy_address", req.ProxyAddr.String()),
		})

	return &types.RegisterProxyResponse{}, nil
}

func (s msgServer) DeactivateProxy(c context.Context, req *types.DeactivateProxyRequest) (*types.DeactivateProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	proxy, _ := s.Keeper.GetProxy(ctx, req.PrincipalAddr)

	if err := s.Keeper.DeactivateProxy(ctx, req.PrincipalAddr); err != nil {
		return nil, sdkerrors.Wrap(types.ErrSnapshot, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeDeactivateProxy),
			sdk.NewAttribute(sdk.AttributeKeySender, req.PrincipalAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, proxy.String()),
		),
	)

	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "deactivate", "proxy"},
		float32(time.Now().Unix()),
		[]metrics.Label{
			telemetry.NewLabel("principal_address", req.PrincipalAddr.String()),
			telemetry.NewLabel("proxy_address", proxy.String()),
		})

	s.Keeper.Logger(ctx).Info(fmt.Sprintf("validator %s has de-activated proxy %s", req.PrincipalAddr, proxy))
	return &types.DeactivateProxyResponse{}, nil
}
