package btc_bridge

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

// For testing purposes only
func NewDummyHandler(k keeper.Keeper, v types.Voter) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrackAddress:
			return dummyHandleMsgTrackAddress(ctx, k, msg)
		case types.MsgVerifyTx:
			return dummyHandleMsgVerifyTx(ctx, k, v, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func dummyHandleMsgTrackAddress(ctx sdk.Context, k keeper.Keeper, msg types.MsgTrackAddress) (*sdk.Result, error) {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", msg.Address))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
		),
	)

	return &sdk.Result{
		Log:    "running without btc bridge",
		Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
		Events: ctx.EventManager().Events(),
	}, nil
}

func dummyHandleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")

	if _, err := chainhash.NewHashFromStr(msg.Tx.TxID); err != nil {
		k.Logger(ctx).Info(err.Error())
		return nil, sdkerrors.Wrap(err, "could not transform Bitcoin transaction ID to hash")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTx, msg.Tx.String()),
		),
	)

	v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: false})
	return &sdk.Result{
		Log:    "running without btc bridge",
		Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
		Events: ctx.EventManager().Events(),
	}, nil
}
