package keeper

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/utils/funcs"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

var _ types.MsgServiceServer = msgServer{}

const allChain = ":all:"

type msgServer struct {
	types.Nexus
	snapshotter types.Snapshotter
	slashing    types.SlashingKeeper
	staking     types.StakingKeeper
	axelarnet   types.AxelarnetKeeper
}

// NewMsgServerImpl returns an implementation of the nexus MsgServiceServer interface for the provided Keeper.
func NewMsgServerImpl(k types.Nexus, snapshotter types.Snapshotter, slashing types.SlashingKeeper, staking types.StakingKeeper, axelarnet types.AxelarnetKeeper) types.MsgServiceServer {
	return msgServer{
		Nexus:       k,
		snapshotter: snapshotter,
		slashing:    slashing,
		staking:     staking,
		axelarnet:   axelarnet,
	}
}

func (s msgServer) RegisterChainMaintainer(c context.Context, req *types.RegisterChainMaintainerRequest) (*types.RegisterChainMaintainerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	validator := s.snapshotter.GetOperator(ctx, req.Sender)
	if validator.Empty() {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", req.Sender.String())
	}

	if s.staking.Validator(ctx, validator) == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator", validator)
	}

	for _, chainStr := range req.Chains {
		chain, ok := s.GetChain(ctx, chainStr)
		if !ok {
			s.Logger(ctx).Error(fmt.Sprintf("%s is not a registered chain", chainStr))
			continue
		}

		if s.axelarnet.IsCosmosChain(ctx, chain.Name) {
			s.Logger(ctx).Error(fmt.Sprintf("'%s' is a cosmos chain, skipping maintainer registration", chain.Name))
			continue
		}
		if s.IsChainMaintainer(ctx, chain, validator) {
			s.Logger(ctx).Info(fmt.Sprintf("'%s' is already a maintainer for chain '%s'", validator.String(), chain.Name))
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChainMaintainer,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueRegister),
				sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
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
				sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
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

func (s msgServer) ActivateChain(c context.Context, req *types.ActivateChainRequest) (*types.ActivateChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if strings.ToLower(req.Chains[0].String()) == allChain {
		for _, chain := range s.GetChains(ctx) {
			s.activateChain(ctx, chain)
		}
	} else {
		for _, chainStr := range req.Chains {
			chain, ok := s.GetChain(ctx, chainStr)
			if !ok {
				return nil, fmt.Errorf("%s is not a registered chain", chainStr)
			}
			s.activateChain(ctx, chain)
		}
	}
	return &types.ActivateChainResponse{}, nil
}

// DeactivateChain handles deactivate chains in case of emergencies
func (s msgServer) DeactivateChain(c context.Context, req *types.DeactivateChainRequest) (*types.DeactivateChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if strings.ToLower(req.Chains[0].String()) == allChain {
		for _, chain := range s.GetChains(ctx) {
			s.deactivateChain(ctx, chain)
		}
	} else {
		for _, chainStr := range req.Chains {
			chain, ok := s.GetChain(ctx, chainStr)
			if !ok {
				return nil, fmt.Errorf("%s is not a registered chain", chainStr)
			}

			s.deactivateChain(ctx, chain)
		}
	}
	return &types.DeactivateChainResponse{}, nil
}

func (s msgServer) activateChain(ctx sdk.Context, chain exported.Chain) {
	if s.IsChainActivated(ctx, chain) {
		s.Logger(ctx).Info(fmt.Sprintf("chain %s already activated", chain.Name))
		return
	}

	// no chain maintainer for cosmos chains
	if !s.axelarnet.IsCosmosChain(ctx, chain.Name) {
		if err := isActivationThresholdMet(ctx, s.Nexus, s.slashing, s.snapshotter, chain); err != nil {
			s.Logger(ctx).Info(fmt.Sprintf("activation threshold is not met for %s due to: %s", chain.Name, err.Error()))
			return
		}
	}

	s.Nexus.ActivateChain(ctx, chain)

	s.Logger(ctx).Info(fmt.Sprintf("chain %s activated", chain.Name))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChain,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueActivated),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
		),
	)
}

func (s msgServer) deactivateChain(ctx sdk.Context, chain exported.Chain) {
	if !s.IsChainActivated(ctx, chain) {
		s.Logger(ctx).Info(fmt.Sprintf("chain %s already deactivated", chain.Name))
		return
	}

	s.Nexus.DeactivateChain(ctx, chain)

	s.Logger(ctx).Info(fmt.Sprintf("chain %s deactivated", chain.Name))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChain,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDeactivated),
			sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
		),
	)
}

func isActivationThresholdMet(ctx sdk.Context, nexus types.Nexus, slashing types.SlashingKeeper, snapshotter types.Snapshotter, chain exported.Chain) error {
	isTombstoned := func(v snapshot.ValidatorI) bool {
		consAdd, err := v.GetConsAddr()
		if err != nil {
			return true
		}

		return slashing.IsTombstoned(ctx, consAdd)
	}

	isProxyActive := func(v snapshot.ValidatorI) bool {
		_, isActive := snapshotter.GetProxy(ctx, v.GetOperator())

		return isActive
	}

	filter := funcs.And(
		snapshot.ValidatorI.IsBonded,
		funcs.Not(snapshot.ValidatorI.IsJailed),
		funcs.Not(isTombstoned),
		isProxyActive,
	)

	params := nexus.GetParams(ctx)

	_, err := snapshotter.CreateSnapshot(
		ctx,
		nexus.GetChainMaintainers(ctx, chain),
		filter,
		snapshot.QuadraticWeightFunc,
		params.ChainActivationThreshold,
	)

	return err
}

func (s msgServer) RegisterAssetFee(c context.Context, req *types.RegisterAssetFeeRequest) (*types.RegisterAssetFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := s.GetChain(ctx, req.FeeInfo.Chain)
	if !ok {
		return nil, fmt.Errorf("%s is not a registered chain", req.FeeInfo.Chain)
	}

	if err := s.RegisterFee(ctx, chain, req.FeeInfo); err != nil {
		return nil, err
	}

	s.Logger(ctx).Info(fmt.Sprintf("registered fee info for asset %s on chain %s", req.FeeInfo.Asset, chain.Name), types.AttributeKeyChain, chain.Name, types.AttributeKeyAsset, req.FeeInfo.Asset)

	return &types.RegisterAssetFeeResponse{}, nil
}
