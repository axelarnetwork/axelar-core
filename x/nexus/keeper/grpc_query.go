package keeper

import (
	"context"
	"fmt"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.QueryServiceServer = Keeper{}

// TransfersForChain returns the transfers for a given chain
func (k Keeper) TransfersForChain(c context.Context, req *types.TransfersForChainRequest) (*types.TransfersForChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := k.GetChain(ctx, req.Chain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.Chain)
	}

	if req.State == nexus.TRANSFER_STATE_UNSPECIFIED {
		return nil, fmt.Errorf("invalid transfer state")
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

// FeeInfo returns the fee info for an asset on a specific chain
func (k Keeper) FeeInfo(c context.Context, req *types.FeeInfoRequest) (*types.FeeInfoResponse, error) {
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

	return &types.FeeInfoResponse{FeeInfo: &feeInfo}, nil
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

	if !k.IsAssetRegistered(ctx, sourceChain, req.Asset) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", req.Asset, sourceChain.Name)
	}

	if !k.IsAssetRegistered(ctx, destinationChain, req.Asset) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", req.Asset, destinationChain.Name)
	}

	asset := sdk.NewCoin(req.Asset, sdk.Int(req.Amount))
	transferFees, feeInfo := k.computeTransferFee(ctx, sourceChain, destinationChain, asset)
	fees := sdk.Uint(transferFees.Amount)

	amount := sdk.ZeroUint()
	if fees.LT(req.Amount) {
		amount = req.Amount.Sub(fees)
	}

	return &types.TransferFeeResponse{Fees: fees, Received: amount, FeeInfo: &feeInfo}, nil
}
