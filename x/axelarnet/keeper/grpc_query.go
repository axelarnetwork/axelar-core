package keeper

import (
	"context"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.QueryServiceServer = queryServer{}

type queryServer struct {
	keeper types.BaseKeeper
	nexus  types.Nexus
}

// NewQueryServerImpl returns an implementation of the axelarnet QueryServiceServer interface for the provided Keeper.
func NewQueryServerImpl(k types.BaseKeeper, n types.Nexus) types.QueryServiceServer {
	return queryServer{
		keeper: k,
		nexus:  n,
	}
}

// DepositAddress returns the deposit address for the provided recipient
func (q queryServer) DepositAddress(c context.Context, req *types.DepositAddressRequest) (*types.DepositAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	recipientChain, ok := q.nexus.GetChain(ctx, req.RecipientChain)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrAxelarnet, "%s is not a registered chain", req.RecipientChain)
	}

	recipient := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}
	address, ok := q.keeper.GetDepositAddress(ctx, recipient)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrAxelarnet, "no deposit address found for %s", recipient.String())
	}

	return &types.DepositAddressResponse{DepositAddr: address.String()}, nil
}
