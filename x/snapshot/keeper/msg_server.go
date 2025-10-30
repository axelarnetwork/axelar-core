package keeper

import (
	"context"
	"fmt"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/hashicorp/go-metrics"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the snapshot MsgServiceServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServiceServer {
	return msgServer{
		Keeper: keeper,
	}
}

func (s msgServer) RegisterProxy(c context.Context, req *types.RegisterProxyRequest) (*types.RegisterProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid sender: %s", err)
	}

	if err := s.ActivateProxy(ctx, sdk.ValAddress(sender), req.ProxyAddr); err != nil {
		return nil, errorsmod.Wrap(types.ErrSnapshot, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeRegisterProxy),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender),
			sdk.NewAttribute(types.AttributeAddress, req.ProxyAddr.String()),
		),
	)

	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "register", "proxy"},
		0,
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
			telemetry.NewLabel("principal_address", req.Sender),
			telemetry.NewLabel("proxy_address", req.ProxyAddr.String()),
		})

	s.Keeper.Logger(ctx).Info(fmt.Sprintf("validator %s registered proxy %s", req.Sender, req.ProxyAddr.String()))
	return &types.RegisterProxyResponse{}, nil
}

func (s msgServer) DeactivateProxy(c context.Context, req *types.DeactivateProxyRequest) (*types.DeactivateProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid sender: %s", err)
	}

	proxy, _ := s.GetProxy(ctx, sdk.ValAddress(sender))

	if err := s.Keeper.DeactivateProxy(ctx, sdk.ValAddress(sender)); err != nil {
		return nil, errorsmod.Wrap(types.ErrSnapshot, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeDeactivateProxy),
			sdk.NewAttribute(sdk.AttributeKeySender, req.Sender),
			sdk.NewAttribute(types.AttributeAddress, proxy.String()),
		),
	)

	telemetry.SetGaugeWithLabels(
		[]string{types.ModuleName, "deactivate", "proxy"},
		0,
		[]metrics.Label{
			telemetry.NewLabel("timestamp", strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
			telemetry.NewLabel("principal_address", req.Sender),
			telemetry.NewLabel("proxy_address", proxy.String()),
		})

	s.Logger(ctx).Info(fmt.Sprintf("validator %s has de-activated proxy %s", req.Sender, proxy.String()))

	return &types.DeactivateProxyResponse{}, nil
}
