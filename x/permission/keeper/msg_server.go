package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns a new msg server instance
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

func (s msgServer) UpdateGovernanceKey(c context.Context, req *types.UpdateGovernanceKeyRequest) (*types.UpdateGovernanceKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := s.getGovAccount(ctx, req.GovernanceKey.Address().Bytes()); ok {
		return nil, fmt.Errorf("account is already registered with a role")
	}

	s.setGovernanceKey(ctx, req.GovernanceKey)
	// delete the existing governance account address
	s.deleteGovAccount(ctx, req.Sender)

	s.setGovAccount(ctx, types.NewGovAccount(req.GovernanceKey.Address().Bytes(), exported.ROLE_ACCESS_CONTROL))

	return &types.UpdateGovernanceKeyResponse{}, nil
}

// RegisterController handles register a controller account
func (s msgServer) RegisterController(c context.Context, req *types.RegisterControllerRequest) (*types.RegisterControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := s.getGovAccount(ctx, req.Controller); ok {
		return nil, fmt.Errorf("account is already registered with a role")
	}

	s.setGovAccount(ctx, types.NewGovAccount(req.Controller, exported.ROLE_CHAIN_MANAGEMENT))

	return &types.RegisterControllerResponse{}, nil
}

// DeregisterController handles delete a controller account
func (s msgServer) DeregisterController(c context.Context, req *types.DeregisterControllerRequest) (*types.DeregisterControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if s.GetRole(ctx, req.Controller) == exported.ROLE_CHAIN_MANAGEMENT {
		s.deleteGovAccount(ctx, req.Controller)
	}

	return &types.DeregisterControllerResponse{}, nil
}
