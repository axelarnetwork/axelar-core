package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

var _ types.QueryServiceServer = Keeper{}

// TransfersForChain returns the transfers for a given chain
func (k Keeper) TransfersForChain(c context.Context, req *types.TransfersForChainRequest) (*types.TransfersForChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := k.GetChain(ctx, req.Chain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.Chain)
	}

	if err := req.State.Validate(); err != nil {
		return nil, err
	}

	transfers, pagination, err := k.GetTransfersForChainPaginated(ctx, chain, req.State, req.Pagination)
	return &types.TransfersForChainResponse{Transfers: transfers, Pagination: pagination}, err
}

// LatestDepositAddress returns the deposit address for the provided recipient
func (k Keeper) LatestDepositAddress(c context.Context, req *types.LatestDepositAddressRequest) (*types.LatestDepositAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	recipientChain, ok := k.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.RecipientChain)
	}

	depositChain, ok := k.GetChain(ctx, req.DepositChain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.DepositChain)
	}

	recipientAddress := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}
	depositAddress, ok := k.getLatestDepositAddress(ctx, depositChain.Name, recipientAddress)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "no deposit address found for recipient %s on chain %s", req.RecipientAddr, req.RecipientChain)
	}

	return &types.LatestDepositAddressResponse{DepositAddr: depositAddress.Address}, nil
}

// Fee returns the fee info for an asset on a specific chain
func (k Keeper) Fee(c context.Context, req *types.FeeRequest) (*types.FeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := k.GetChain(ctx, req.Chain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.Chain)
	}

	if !k.IsAssetRegistered(ctx, chain, req.Asset) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", req.Asset, chain.Name)
	}

	feeInfo, ok := k.GetFeeInfo(ctx, chain, req.Asset)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "no fee info registered for asset %s on chain %s", req.Asset, chain.Name)
	}

	return &types.FeeResponse{FeeInfo: &feeInfo}, nil
}

// TransferFee returns the transfer fee for a cross chain transfer
func (k Keeper) TransferFee(c context.Context, req *types.TransferFeeRequest) (*types.TransferFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sourceChain, ok := k.GetChain(ctx, req.SourceChain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.SourceChain)
	}

	destinationChain, ok := k.GetChain(ctx, req.DestinationChain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.DestinationChain)
	}

	amount, err := req.GetAmount()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid amount")
	}

	if !k.IsAssetRegistered(ctx, sourceChain, amount.Denom) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", amount.Denom, sourceChain.Name)
	}

	if !k.IsAssetRegistered(ctx, destinationChain, amount.Denom) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", amount.Denom, destinationChain.Name)
	}

	if amount.IsNegative() {
		return nil, fmt.Errorf("amount cannot be negative")
	}

	fee, err := k.ComputeTransferFee(ctx, sourceChain, destinationChain, amount)
	if err != nil {
		return nil, err
	}

	return &types.TransferFeeResponse{Fee: fee}, nil
}
