package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/slices"
)

var _ types.QueryServiceServer = Querier{}

// Querier implements the grpc queries for the nexus module
type Querier struct {
	keeper    Keeper
	axelarnet types.AxelarnetKeeper
}

// NewGRPCQuerier creates a new nexus Querier
func NewGRPCQuerier(k Keeper, a types.AxelarnetKeeper) Querier {
	return Querier{
		keeper:    k,
		axelarnet: a,
	}
}

// Params returns the reward module params
func (q Querier) Params(c context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params := q.keeper.GetParams(ctx)

	return &types.ParamsResponse{
		Params: params,
	}, nil
}

// TransfersForChain returns the transfers for a given chain
func (q Querier) TransfersForChain(c context.Context, req *types.TransfersForChainRequest) (*types.TransfersForChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.Chain)
	}

	if err := req.State.Validate(); err != nil {
		return nil, err
	}

	transfers, pagination, err := q.keeper.GetTransfersForChainPaginated(ctx, chain, req.State, req.Pagination)
	return &types.TransfersForChainResponse{Transfers: transfers, Pagination: pagination}, err
}

// LatestDepositAddress returns the deposit address for the provided recipient
func (q Querier) LatestDepositAddress(c context.Context, req *types.LatestDepositAddressRequest) (*types.LatestDepositAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	recipientChain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.RecipientChain))
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.RecipientChain)
	}

	depositChain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.DepositChain))
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.DepositChain)
	}

	recipientAddress := nexus.CrossChainAddress{Chain: recipientChain, Address: req.RecipientAddr}
	depositAddress, ok := q.keeper.getLatestDepositAddress(ctx, depositChain.Name, recipientAddress)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "no deposit address found for recipient %s on chain %s", req.RecipientAddr, req.RecipientChain)
	}

	return &types.LatestDepositAddressResponse{DepositAddr: depositAddress.Address}, nil
}

// FeeInfo returns the fee info for an asset on a specific chain
func (q Querier) FeeInfo(c context.Context, req *types.FeeInfoRequest) (*types.FeeInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.Chain)
	}

	if !q.keeper.IsAssetRegistered(ctx, chain, req.Asset) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", req.Asset, chain.Name)
	}

	feeInfo := q.keeper.GetFeeInfo(ctx, chain, req.Asset)

	return &types.FeeInfoResponse{FeeInfo: &feeInfo}, nil
}

// TransferFee returns the transfer fee for a cross chain transfer
func (q Querier) TransferFee(c context.Context, req *types.TransferFeeRequest) (*types.TransferFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	sourceChain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.SourceChain))
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.SourceChain)
	}

	destinationChain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.DestinationChain))
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered chain", req.DestinationChain)
	}

	amount, err := req.GetAmount()
	if err != nil {
		return nil, err
	}

	if !q.keeper.IsAssetRegistered(ctx, sourceChain, amount.Denom) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", amount.Denom, sourceChain.Name)
	}

	if !q.keeper.IsAssetRegistered(ctx, destinationChain, amount.Denom) {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "%s is not a registered asset on chain %s", amount.Denom, destinationChain.Name)
	}

	if amount.IsNegative() {
		return nil, fmt.Errorf("amount cannot be negative")
	}

	// When source chain is another cosmos chain, use axelarnet for fee info where deposit address is generated on
	feeCalcSourceChain := sourceChain
	if q.axelarnet.IsCosmosChain(ctx, feeCalcSourceChain.Name) {
		feeCalcSourceChain = exported.Axelarnet
	}

	fee, err := q.keeper.ComputeTransferFee(ctx, feeCalcSourceChain, destinationChain, amount)
	if err != nil {
		return nil, err
	}

	return &types.TransferFeeResponse{Fee: fee}, nil
}

// Chains returns the chains registered on the network
func (q Querier) Chains(c context.Context, req *types.ChainsRequest) (*types.ChainsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chains := q.keeper.GetChains(ctx)

	switch req.Status {
	case types.Activated:
		chains = slices.Filter(chains, func(chain nexus.Chain) bool { return q.keeper.IsChainActivated(ctx, chain) })
	case types.Deactivated:
		chains = slices.Filter(chains, func(chain nexus.Chain) bool { return !q.keeper.IsChainActivated(ctx, chain) })
	}

	chainNames := slices.Map(chains, nexus.Chain.GetName)

	return &types.ChainsResponse{Chains: chainNames}, nil
}

// Assets returns the registered assets of a chain
func (q Querier) Assets(c context.Context, req *types.AssetsRequest) (*types.AssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, fmt.Errorf("chain %s not found", req.Chain)
	}

	chainState, ok := q.keeper.getChainState(ctx, chain)
	if !ok {
		return nil, fmt.Errorf("chain state not found for %s", chain.Name)
	}

	assets := make([]string, len(chainState.Assets))
	for i, asset := range chainState.Assets {
		assets[i] = asset.Denom
	}

	return &types.AssetsResponse{Assets: assets}, nil
}

// ChainState returns the chain state in the network
func (q Querier) ChainState(c context.Context, req *types.ChainStateRequest) (*types.ChainStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, fmt.Errorf("chain %s not found", req.Chain)
	}

	chainState, ok := q.keeper.getChainState(ctx, chain)
	if !ok {
		return nil, fmt.Errorf("chain state not found for %s", chain.Name)
	}

	return &types.ChainStateResponse{State: chainState}, nil
}

// ChainsByAsset returns all chains that an asset is registered on
func (q Querier) ChainsByAsset(c context.Context, req *types.ChainsByAssetRequest) (*types.ChainsByAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if err := sdk.ValidateDenom(req.Asset); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid asset")
	}

	chains := q.keeper.GetChains(ctx)
	var chainNames []nexus.ChainName

	for _, chain := range chains {
		if q.keeper.IsAssetRegistered(ctx, chain, req.Asset) {
			chainNames = append(chainNames, chain.Name)
		}
	}

	return &types.ChainsByAssetResponse{Chains: chainNames}, nil
}

// RecipientAddress returns the recipient address for a given deposit address
func (q Querier) RecipientAddress(c context.Context, req *types.RecipientAddressRequest) (*types.RecipientAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.DepositChain))
	if !ok {
		return nil, fmt.Errorf("chain %s not found", req.DepositChain)
	}

	depositAddress := nexus.CrossChainAddress{Chain: chain, Address: req.DepositAddr}

	linkedAddresses, ok := q.keeper.getLinkedAddresses(ctx, depositAddress)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrNexus, "no recipient address found for deposit address %s on chain %s", req.DepositAddr, req.DepositChain)
	}

	return &types.RecipientAddressResponse{
		RecipientAddr:  linkedAddresses.RecipientAddress.Address,
		RecipientChain: linkedAddresses.RecipientAddress.Chain.Name.String(),
	}, nil
}

// ChainMaintainers returns the chain maintainers for a given chain
func (q Querier) ChainMaintainers(c context.Context, req *types.ChainMaintainersRequest) (*types.ChainMaintainersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, status.Errorf(codes.NotFound, "invalid chain %s", req.Chain)
	}

	maintainers := q.keeper.GetChainMaintainers(ctx, chain)

	return &types.ChainMaintainersResponse{Maintainers: maintainers}, nil
}

// TransferRateLimit queries the transfer rate limit for a given chain and asset
func (q Querier) TransferRateLimit(c context.Context, req *types.TransferRateLimitRequest) (*types.TransferRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	chain, ok := q.keeper.GetChain(ctx, nexus.ChainName(req.Chain))
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, fmt.Errorf("chain %s not found", req.Chain).Error())
	}

	rateLimit, found := q.keeper.getRateLimit(ctx, chain.Name, req.Asset)
	if !found {
		return &types.TransferRateLimitResponse{}, nil
	}

	fromDirectionEpoch := q.keeper.getCurrentTransferEpoch(ctx, chain.Name, req.Asset, nexus.TransferDirectionFrom, rateLimit.Window)
	toDirectionEpoch := q.keeper.getCurrentTransferEpoch(ctx, chain.Name, req.Asset, nexus.TransferDirectionTo, rateLimit.Window)

	// time left = (epoch + 1) * window - current time
	timeLeft := time.Duration(int64(fromDirectionEpoch.Epoch+1)*int64(rateLimit.Window) - ctx.BlockTime().UnixNano())

	return &types.TransferRateLimitResponse{
		TransferRateLimit: &types.TransferRateLimit{
			Limit:    rateLimit.Limit.Amount,
			Window:   rateLimit.Window,
			Incoming: fromDirectionEpoch.Amount.Amount,
			Outgoing: toDirectionEpoch.Amount.Amount,
			TimeLeft: timeLeft,
			From:     fromDirectionEpoch.Amount.Amount,
			To:       toDirectionEpoch.Amount.Amount,
		},
	}, nil
}

// Message queries the general message for a given message ID
func (q Querier) Message(c context.Context, req *types.MessageRequest) (*types.MessageResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	id := req.ID

	msg, found := q.keeper.GetMessage(ctx, id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "message not found: %s", id)
	}

	return &types.MessageResponse{
		Message: msg,
	}, nil
}
