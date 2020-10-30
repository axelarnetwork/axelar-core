package btc_bridge

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

const (
	Satoshi = 1
	Bitcoin = 100_000_000 * Satoshi
)

// This handler is only needed to avoid nil errors, all bridge messages are routed through the axelar module
func NewHandler(k keeper.Keeper, v types.Voter, rpc *rpcclient.Client) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrackAddress:
			return handleMsgTrackAddress(ctx, k, rpc, msg)
		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, v, rpc, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgTrackAddress(ctx sdk.Context, k keeper.Keeper, rpc *rpcclient.Client, msg types.MsgTrackAddress) (*sdk.Result, error) {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address %v", msg.Address))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeTrackAddress),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
		),
	)

	if rpc == nil {
		k.Logger(ctx).Debug("running without btc bridge")
		return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	}

	future := rpc.ImportAddressAsync(msg.Address)
	k.SetTrackedAddress(ctx, msg.Address)

	// ctx might not be valid anymore when future returns, so closing over logger to be safe
	go func(logger log.Logger) {
		if err := future.Receive(); err != nil {
			logger.Error(fmt.Sprintf("Could not track address %v", msg.Address))
		} else {
			logger.Debug(fmt.Sprintf("successfully tracked all past transaction for address %v", msg.Address))
		}
	}(k.Logger(ctx))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, v types.Voter, rpc *rpcclient.Client, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeVerifyTx),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeTx, msg.Tx.String()),
		),
	)

	if rpc == nil {
		v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: false})
		return &sdk.Result{
			Log:    "running without btc bridge",
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	// This is the only check where the input can be faulty, therefore do not vote on error and discard tx
	hash, err := chainhash.NewHashFromStr(msg.Tx.TxID)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		return nil, sdkerrors.Wrap(err, "could not transform Bitcoin transaction ID to hash")
	}

	btcTxResult, err := rpc.GetTransaction(hash)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: false})
		return &sdk.Result{
			Log:    sdkerrors.Wrap(err, "could not retrieve Bitcoin transaction").Error(),
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	actualAmount, err := btcutil.NewAmount(btcTxResult.Amount)
	if err != nil {
		k.Logger(ctx).Info(err.Error())
		v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: false})
		return &sdk.Result{
			Log:    sdkerrors.Wrap(err, "could not parse transaction amount of the Bitcoin response").Error(),
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	expectedAmount := msg.Tx.Amount.Amount

	isEqual := btcTxResult.TxID == msg.Tx.TxID &&
		amountEquals(expectedAmount, actualAmount) &&
		btcTxResult.Confirmations >= k.GetConfirmationHeight(ctx)

	if !isEqual {
		k.Logger(ctx).Debug(fmt.Sprintf(
			"txID:%s\nbtcTxId:%s\ntx amount:%s\nbtc Amount:%v",
			msg.Tx.TxID,
			btcTxResult.TxID,
			msg.Tx.Amount.String(),
			actualAmount,
		))
		v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: false})
		return &sdk.Result{
			Log:    sdkerrors.Wrap(err, "expected transaction differs from actual transaction on Bitcoin").Error(),
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: true})
	return &sdk.Result{
		Log:    sdkerrors.Wrap(err, "successfully verified transaction").Error(),
		Data:   k.Codec().MustMarshalBinaryLengthPrefixed(true),
		Events: ctx.EventManager().Events(),
	}, nil
}

func amountEquals(expectedAmount sdk.Dec, actualAmount btcutil.Amount) bool {
	return (expectedAmount.IsInteger() && satoshiEquals(expectedAmount, actualAmount)) ||
		btcEquals(expectedAmount, actualAmount)
}

func satoshiEquals(satoshiAmount sdk.Dec, verifiedAmount btcutil.Amount) bool {
	return satoshiAmount.IsInt64() && btcutil.Amount(satoshiAmount.Int64()) == verifiedAmount
}

func btcEquals(btcAmount sdk.Dec, verifiedAmount btcutil.Amount) bool {
	return btcutil.Amount(btcAmount.MulInt64(Bitcoin).Int64()) == verifiedAmount
}
