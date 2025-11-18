package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the nexus module
type Querier struct {
	keeper Keeper
}

// NewGRPCQuerier creates a new nexus Querier
func NewGRPCQuerier(k Keeper) Querier {
	return Querier{
		keeper: k,
	}
}

// Params returns the reward module params
func (q Querier) Params(c context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := q.keeper.GetParams(ctx)

	return &types.ParamsResponse{
		Params: params,
	}, nil
}

func (q Querier) OperatorByProxy(c context.Context, req *types.OperatorByProxyRequest) (*types.OperatorByProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	proxyAddr, err := sdk.AccAddressFromBech32(req.ProxyAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrSnapshot, "invalid proxy address")
	}

	operator := q.keeper.GetOperator(ctx, proxyAddr)
	if operator == nil {
		return nil, errorsmod.Wrap(types.ErrSnapshot, "no operator associated to the proxy address")
	}

	return &types.OperatorByProxyResponse{
		OperatorAddress: operator.String(),
	}, nil
}

func (q Querier) ProxyByOperator(c context.Context, req *types.ProxyByOperatorRequest) (*types.ProxyByOperatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	operatorAddr, err := sdk.ValAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrSnapshot, "invalid operator address")
	}

	proxy, active := q.keeper.GetProxy(ctx, operatorAddr)
	if proxy == nil {
		return nil, errorsmod.Wrap(types.ErrSnapshot, "no proxy set for operator address")
	}

	status := types.Inactive
	if active {
		status = types.Active
	}

	return &types.ProxyByOperatorResponse{
		ProxyAddress: proxy.String(),
		Status:       status,
	}, nil

}
