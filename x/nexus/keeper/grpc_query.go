package keeper

import (
	"context"
	"strings"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
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

	var state exported.TransferState
	switch strings.ToLower(req.State) {
	case "pending":
		state = exported.Pending
	case "archived":
		state = exported.Archived
	default:
		return nil, sdkerrors.Wrapf(types.ErrNexus, "unknown transfer state '%s'", req.State)
	}

	transfers := k.GetTransfersForChain(ctx, chain, state)
	transfersResp := make([]types.TransfersForChainResponse_Transfer, 0)

	for _, transfer := range transfers {
		transfersResp = append(transfersResp, types.TransfersForChainResponse_Transfer{
			ID:        transfer.ID.String(),
			Recipient: transfer.Recipient.Address,
			Asset:     transfer.Asset.String(),
		})
	}

	return &types.TransfersForChainResponse{Transfers: transfersResp, Count: int32(len(transfers))}, nil
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
