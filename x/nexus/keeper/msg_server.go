package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.Nexus
	snapshotter types.Snapshotter
}

// NewMsgServerImpl returns an implementation of the nexus MsgServiceServer interface for the provided Keeper.
func NewMsgServerImpl(k types.Nexus, snapshotter types.Snapshotter) types.MsgServiceServer {
	return msgServer{
		Nexus:       k,
		snapshotter: snapshotter,
	}
}

func (s msgServer) RegisterChainMaintainer(c context.Context, req *types.RegisterChainMaintainerRequest) (*types.RegisterChainMaintainerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	validator := s.snapshotter.GetOperator(ctx, req.Sender)
	if validator.Empty() {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	for _, chainStr := range req.Chains {
		chain, ok := s.GetChain(ctx, chainStr)
		if !ok {
			return nil, fmt.Errorf("%s is not a registered chain", chainStr)
		}

		if s.IsChainMaintainer(ctx, chain, validator) {
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChainMaintainer,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueRegister),
				sdk.NewAttribute(types.AttributeKeyChain, chainStr),
				sdk.NewAttribute(types.AttributeKeyChainMaintainerAddress, validator.String()),
			),
		)

		s.Logger(ctx).Info(fmt.Sprintf("validator %s registered maintainer for chain %s", validator.String(), chain.Name))
		if err := s.AddChainMaintainer(ctx, chain, validator); err != nil {
			return nil, err
		}
	}

	return &types.RegisterChainMaintainerResponse{}, nil
}

func (s msgServer) DeregisterChainMaintainer(c context.Context, req *types.DeregisterChainMaintainerRequest) (*types.DeregisterChainMaintainerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	validator := s.snapshotter.GetOperator(ctx, req.Sender)
	if validator.Empty() {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	for _, chainStr := range req.Chains {
		chain, ok := s.GetChain(ctx, chainStr)
		if !ok {
			return nil, fmt.Errorf("%s is not a registered chain", chainStr)
		}

		if !s.IsChainMaintainer(ctx, chain, validator) {
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChainMaintainer,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDeregister),
				sdk.NewAttribute(types.AttributeKeyChain, chainStr),
				sdk.NewAttribute(types.AttributeKeyChainMaintainerAddress, validator.String()),
			),
		)

		s.Logger(ctx).Info(fmt.Sprintf("validator %s deregistered maintainer for chain %s", validator.String(), chain.Name))
		if err := s.RemoveChainMaintainer(ctx, chain, validator); err != nil {
			return nil, err
		}
	}

	return &types.DeregisterChainMaintainerResponse{}, nil
}
