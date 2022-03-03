package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc querier
type Querier struct {
	keeper BaseKeeper
	nexus  types.Nexus
}

// NewGRPCQuerier returns a new Querier
func NewGRPCQuerier(k BaseKeeper, n types.Nexus) Querier {
	return Querier{
		keeper: k,
		nexus:  n,
	}
}

// BurnerInfo implements the burner info grpc query
func (q Querier) BurnerInfo(c context.Context, req *types.BurnerInfoRequest) (*types.BurnerInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	iter := q.keeper.getBaseStore(ctx).Iterator(subspacePrefix)
	defer utils.CloseLogError(iter, q.keeper.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		ck := q.keeper.ForChain(string(iter.Value()))
		burnerInfo := ck.GetBurnerInfo(ctx, req.Address)
		if burnerInfo != nil {
			return &types.BurnerInfoResponse{Chain: ck.GetParams(ctx).Chain, BurnerInfo: burnerInfo}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "unknown address")
}

// ConfirmationHeight implements the confirmation height grpc query
func (q Querier) ConfirmationHeight(c context.Context, req *types.ConfirmationHeightRequest) (*types.ConfirmationHeightResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	_, ok := q.nexus.GetChain(ctx, req.Chain)
	if !ok {
		return nil, status.Error(codes.NotFound, "unknown chain")

	}

	ck := q.keeper.ForChain(string(req.Chain))
	height, ok := ck.GetRequiredConfirmationHeight(ctx)
	if !ok {
		return nil, status.Error(codes.NotFound, "could not get confirmation height")
	}

	return &types.ConfirmationHeightResponse{Height: height}, nil
}

// DepositState fetches the state of a deposit confirmation using a grpc query
func (q Querier) DepositState(c context.Context, req *types.DepositStateRequest) (*types.DepositStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	ck := q.keeper.ForChain(req.Chain)

	s, log, code := queryDepositState(ctx, ck, q.nexus, req.Params)
	if code != codes.OK {
		return nil, status.Error(code, log)
	}

	return &types.DepositStateResponse{Status: s}, nil
}

// PendingCommands implements the pending commands query
func (q Querier) PendingCommands(c context.Context, req *types.PendingCommandsRequest) (*types.PendingCommandsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if !q.keeper.HasChain(ctx, req.Chain) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("chain %s not found", req.Chain))
	}

	ck := q.keeper.ForChain(req.Chain)

	pendingCommands, errLog, code := queryPendingCommands(ctx, ck, q.nexus)
	if code != codes.OK {
		return nil, status.Error(codes.NotFound, errLog)
	}

	return &pendingCommands, nil
}
