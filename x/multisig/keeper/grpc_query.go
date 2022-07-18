package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	keyID, _ := q.keeper.GetCurrentKeyID(ctx, nexus.ChainName(req.Chain))

	return &types.KeyIDResponse{KeyID: keyID}, nil
}

// NextKeyID returns the key ID assigned for the next rotation on a given chain and empty if none is assigned
func (q Querier) NextKeyID(c context.Context, req *types.NextKeyIDRequest) (*types.NextKeyIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keyID, _ := q.keeper.GetNextKeyID(ctx, nexus.ChainName(req.Chain))

	return &types.NextKeyIDResponse{KeyID: keyID}, nil
}
