package btc_bridge

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

func NewHandler(k keeper.Keeper, v types.Voter, b types.Bridge) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrackAddress:
			return handleMsgTrackAddress(ctx, k, b, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, b, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgTrackAddress(ctx sdk.Context, k keeper.Keeper, b types.Bridge, msg types.MsgTrackAddress) (*sdk.Result, error) {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", msg.Address))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
		),
	)

	// Importing an address takes a long time, therefore it cannot be done in the critical path.
	// ctx might not be valid anymore when err is returned, so closing over logger to be safe
	go func(logger log.Logger) {
		if err := b.TrackAddress(msg.Address); err != nil {
			logger.Error(fmt.Sprintf("Could not track address %v", msg.Address))
		} else {
			logger.Debug(fmt.Sprintf("successfully tracked all past transaction for address %v", msg.Address))
		}
	}(k.Logger(ctx))

	k.SetTrackedAddress(ctx, msg.Address)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, b types.Bridge, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTx, msg.Tx.String()),
		),
	)

	// Validators can decide this check deterministically, so no need to go through 2nd layer consensus to discard
	_, err := chainhash.NewHashFromStr(msg.Tx.TxID)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		return nil, sdkerrors.Wrap(err, "could not transform Bitcoin transaction ID to hash")
	}

	if err = b.VeriyfyTx(msg.Tx); err != nil {
		v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: false})
		k.Logger(ctx).Debug(sdkerrors.Wrapf(err,
			"expected transaction (%s) could not be verified", msg.Tx.String()).Error())
		return &sdk.Result{
			Log:    err.Error(),
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
	} else {
		v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: true})
		return &sdk.Result{
			Log:    "successfully verified transaction",
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(true),
			Events: ctx.EventManager().Events(),
		}, nil
	}
}
