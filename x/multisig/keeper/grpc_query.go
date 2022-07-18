package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the multisig module
type Querier struct {
	keeper types.Keeper
	staker types.Staker
}

// NewGRPCQuerier creates a new multisig Querier
func NewGRPCQuerier(k types.Keeper, s types.Staker) Querier {
	return Querier{
		keeper: k,
		staker: s,
	}
}

// KeyID returns the key ID assigned to a given chain
func (q Querier) KeyID(c context.Context, req *types.KeyIDRequest) (*types.KeyIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keyID, ok := q.keeper.GetCurrentKeyID(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("key id not found for chain [%s]", req.Chain)).Error())
	}

	return &types.KeyIDResponse{KeyID: keyID}, nil
}

// NextKeyID returns the key ID assigned for the next rotation on a given chain and empty if none is assigned
func (q Querier) NextKeyID(c context.Context, req *types.NextKeyIDRequest) (*types.NextKeyIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keyID, ok := q.keeper.GetNextKeyID(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrMultisig, fmt.Sprintf("next key id not found for chain [%s]", req.Chain)).Error())
	}

	return &types.NextKeyIDResponse{KeyID: keyID}, nil
}

// AssignableKey returns true if no key is assigned for rotation on the given chain, and false otherwise
func (q Querier) AssignableKey(c context.Context, req *types.AssignableKeyRequest) (*types.AssignableKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, assigned := q.keeper.GetNextKeyID(ctx, nexus.ChainName(req.Chain))

	return &types.AssignableKeyResponse{Assignable: !assigned}, nil
}
