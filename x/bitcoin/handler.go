package bitcoin

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

// NewHandler creates an sdk.Handler for all bitcoin type messages
func NewHandler(k types.BTCKeeper, v types.Voter, signer types.Signer, n types.Nexus, snapshotter types.Snapshotter) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, signer, n, v, snapshotter)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.LinkRequest:
			res, err := server.Link(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err == nil {
				result.Log = fmt.Sprintf("successfully linked {%s} and {%s}", res.DepositAddr, msg.RecipientAddr)
			}
			return result, err
		case *types.ConfirmOutpointRequest:
			res, err := server.ConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.VoteConfirmOutpointRequest:
			res, err := server.VoteConfirmOutpoint(sdk.WrapSDKContext(ctx), msg)
			if err == nil {
				k.Logger(ctx).Debug(res.Status)
			}
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreatePendingTransfersTxRequest:
			res, err := server.CreatePendingTransfersTx(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreateMasterTxRequest:
			res, err := server.CreateMasterTx(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.CreateRescueTxRequest:
			res, err := server.CreateRescueTx(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.SignTxRequest:
			res, err := server.SignTx(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.SubmitExternalSignatureRequest:
			res, err := server.SubmitExternalSignature(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			k.Logger(ctx).Debug(err.Error())
			return nil, sdkerrors.Wrap(types.ErrBitcoin, err.Error())
		}
		return res, nil
	}
}
