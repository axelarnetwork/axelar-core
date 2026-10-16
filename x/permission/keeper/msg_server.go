package keeper

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

type msgServer struct {
	Keeper
	upgrade types.UpgradeKeeper
}

// NewMsgServerImpl returns a new msg server instance
func NewMsgServerImpl(keeper Keeper, upgrade types.UpgradeKeeper) types.MsgServer {
	return msgServer{Keeper: keeper, upgrade: upgrade}
}

func (s msgServer) UpdateGovernanceKey(c context.Context, req *types.UpdateGovernanceKeyRequest) (*types.UpdateGovernanceKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := s.getGovAccount(ctx, req.GovernanceKey.Address().Bytes()); ok {
		return nil, fmt.Errorf("account is already registered with a role")
	}

	s.setGovernanceKey(ctx, req.GovernanceKey)

	// delete the existing governance account address
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid sender: %s", err)
	}
	s.deleteGovAccount(ctx, sender)

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

func (s msgServer) UpdateParams(c context.Context, req *types.UpdateParamsRequest) (*types.UpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := req.Params.Validate(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	s.setParams(ctx, req.Params)
	return &types.UpdateParamsResponse{}, nil
}

// Pause pauses the chain at the requested level, or changes the level of an ongoing pause
func (s msgServer) Pause(c context.Context, req *types.PauseRequest) (*types.PauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if s.IsPauseSkipped(ctx) {
		return nil, types.ErrPauseSkipped
	}

	previous := s.GetPauseLevel(ctx)
	if previous == req.Level {
		return nil, fmt.Errorf("chain is already paused at level %s", req.Level)
	}

	s.setPauseLevel(ctx, req.Level)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeEmergencyPause,
		sdk.NewAttribute(types.AttributeKeySigner, req.Sender),
		sdk.NewAttribute(types.AttributeKeyReason, req.Reason),
		sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatInt(ctx.BlockHeight(), 10)),
		sdk.NewAttribute(types.AttributeKeyPauseLevel, req.Level.String()),
		sdk.NewAttribute(types.AttributeKeyPreviousPauseLevel, previous.String()),
	))
	s.Logger(ctx).Info("emergency pause", "level", req.Level.String(), "previous_level", previous.String(), "signer", req.Sender, "reason", req.Reason)

	return &types.PauseResponse{}, nil
}

// Unpause lifts the emergency pause
func (s msgServer) Unpause(c context.Context, req *types.UnpauseRequest) (*types.UnpauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	previous := s.GetPauseLevel(ctx)
	if !previous.IsPaused() {
		return nil, fmt.Errorf("chain is not paused")
	}

	s.setPauseLevel(ctx, exported.PAUSE_LEVEL_UNSPECIFIED)

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeEmergencyUnpause,
		sdk.NewAttribute(types.AttributeKeySigner, req.Sender),
		sdk.NewAttribute(types.AttributeKeyReason, req.Reason),
		sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatInt(ctx.BlockHeight(), 10)),
		sdk.NewAttribute(types.AttributeKeyPreviousPauseLevel, previous.String()),
	))
	s.Logger(ctx).Info("emergency unpause", "previous_level", previous.String(), "signer", req.Sender, "reason", req.Reason)

	return &types.UnpauseResponse{}, nil
}

// HaltChain schedules an upgrade plan at the next block height.
func (s msgServer) HaltChain(c context.Context, req *types.HaltChainRequest) (*types.HaltChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if s.IsPauseSkipped(ctx) {
		return nil, types.ErrPauseSkipped
	}

	pending, err := s.upgrade.GetUpgradePlan(ctx)
	hasPending := err == nil
	switch {
	case err != nil && !errors.Is(err, upgradetypes.ErrNoUpgradePlanFound):
		return nil, err
	case hasPending && !req.OverridePendingUpgrade:
		return nil, fmt.Errorf("upgrade plan %s is already scheduled at height %d, set override_pending_upgrade to replace it", pending.Name, pending.Height)
	}

	height := ctx.BlockHeight() + 1
	plan := upgradetypes.Plan{
		Name:   exported.EmergencyHaltPlanName(height),
		Height: height,
		Info:   req.Reason,
	}

	if err := s.upgrade.ScheduleUpgrade(ctx, plan); err != nil {
		return nil, err
	}

	attributes := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeySigner, req.Sender),
		sdk.NewAttribute(types.AttributeKeyReason, req.Reason),
		sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatInt(ctx.BlockHeight(), 10)),
		sdk.NewAttribute(types.AttributeKeyHaltHeight, strconv.FormatInt(height, 10)),
		sdk.NewAttribute(types.AttributeKeyPlanName, plan.Name),
	}
	if hasPending {
		attributes = append(attributes,
			sdk.NewAttribute(types.AttributeKeyReplacedPlanName, pending.Name),
			sdk.NewAttribute(types.AttributeKeyReplacedPlanHeight, strconv.FormatInt(pending.Height, 10)),
			sdk.NewAttribute(types.AttributeKeyReplacedPlanInfo, pending.Info),
		)
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeEmergencyHalt, attributes...))
	s.Logger(ctx).Info("emergency halt scheduled", "halt_height", height, "plan", plan.Name, "signer", req.Sender, "reason", req.Reason, "replaced_plan", pending.Name)

	return &types.HaltChainResponse{HaltHeight: height, PlanName: plan.Name}, nil
}
