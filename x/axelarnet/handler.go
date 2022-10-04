package axelarnet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

// NewHandler returns the handler of the Cosmos module
func NewHandler(k keeper.Keeper, n types.Nexus, b types.BankKeeper, a types.AccountKeeper, ibcK keeper.IBCKeeper) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, n, b, a, ibcK)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.LinkRequest:
			res, err := server.Link(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully linked deposit %s to recipient %s", res.DepositAddr, msg.RecipientAddr)
			}
			return result, err
		case *types.ConfirmDepositRequest:
			res, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully confirmed deposit to {%s}", msg.DepositAddress.String())
			}
			return result, err
		case *types.ExecutePendingTransfersRequest:
			res, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = "successfully executed pending transfers"
			}
			return result, err
		case *types.RegisterIBCPathRequest:
			res, err := server.RegisterIBCPath(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully registered chain %s with path %s", msg.Chain, msg.Path)
			}
			return result, err
		case *types.AddCosmosBasedChainRequest:
			res, err := server.AddCosmosBasedChain(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully added chain %s", msg.CosmosChain)
			}
			return result, err
		case *types.RegisterAssetRequest:
			res, err := server.RegisterAsset(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully registered asset %s to chain %s", msg.Asset.Denom, msg.Chain)
			}
			return result, err
		case *types.RouteIBCTransfersRequest:
			res, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = "successfully executed pending transfers"
			}
			return result, err
		case *types.RegisterFeeCollectorRequest:
			res, err := server.RegisterFeeCollector(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			return result, err
		case *types.RetryIBCTransferRequest:
			res, err := server.RetryIBCTransfer(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			return result, err
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrAxelarnet, err.Error())
		}
		return res, nil
	}
}
