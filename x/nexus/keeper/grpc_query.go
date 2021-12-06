package keeper

import (
	"context"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.QueryServiceServer = Keeper{}

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

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}
	key := depositAddrPrefix.Append(utils.LowerCaseKey(recipient.String())).Append(utils.LowerCaseKey(depositChain.Name))
	bz := k.getStore(ctx).GetRaw(key)
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "no deposit address found for %s", recipient.String())
	}

	return &types.LatestDepositAddressResponse{DepositAddr: string(bz)}, nil
}
