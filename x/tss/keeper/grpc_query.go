package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the tss module
type Querier struct {
	keeper types.TSSKeeper
	nexus  types.Nexus
}

// NewGRPCQuerier creates a new tss Querier
func NewGRPCQuerier(k types.TSSKeeper, n types.Nexus) Querier {
	return Querier{
		keeper: k,
		nexus:  n,
	}
}

// NextKeyID returns the key ID assigned for the next rotation on a given chain and for the given key role
func (q Querier) NextKeyID(c context.Context, req *types.NextKeyIDRequest) (*types.NextKeyIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrTss, fmt.Sprintf("chain [%s] not found", req.Chain)).Error())
	}

	keyID, ok := q.keeper.GetNextKeyID(ctx, chain, req.KeyRole)
	if !ok {
		return nil, status.Error(codes.OK, fmt.Errorf("no next key assigned for key role [%s] on chain [%s]", req.KeyRole.SimpleString(), chain.Name).Error())
	}

	return &types.NextKeyIDResponse{KeyID: keyID}, nil
}

// AssignableKey returns true if there is assign
func (q Querier) AssignableKey(c context.Context, req *types.AssignableKeyRequest) (*types.AssignableKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, status.Error(codes.NotFound, sdkerrors.Wrap(types.ErrTss, fmt.Sprintf("chain [%s] not found", req.Chain)).Error())
	}

	_, assigned := q.keeper.GetNextKeyID(ctx, chain, req.KeyRole)

	return &types.AssignableKeyResponse{Assignable: !assigned}, nil
}
