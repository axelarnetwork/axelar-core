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
	Satoshi int64 = 1
	Bitcoin       = 100_000_000 * Satoshi
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
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
		),
	)

	if rpc == nil {
		return &sdk.Result{
			Log:    "running without btc bridge",
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
	}

	// Importing an address takes a long time, therefore it cannot be done in the critical path
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

	// This is the only check where the input can be faulty, therefore do not vote on error and discard tx
	hash, err := chainhash.NewHashFromStr(msg.Tx.TxID)
	if err != nil {
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

	if rpc == nil {
		v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: false})
		return &sdk.Result{
			Log:    "running without btc bridge",
			Data:   k.Codec().MustMarshalBinaryLengthPrefixed(false),
			Events: ctx.EventManager().Events(),
		}, nil
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

	v.SetFutureVote(ctx, exported.FutureVote{Tx: msg.Tx, LocalAccept: isEqual})

	var logMsg string
	if isEqual {
		logMsg = "successfully verified transaction"
	} else {
		logMsg = fmt.Sprintf(
			"expected transaction differs from actual transaction on Bitcoin:\n"+
				"expected txID:%s, amount:%s\n"+
				"actual txId:%s, amount:%v",
			msg.Tx.TxID, msg.Tx.Amount.String(), btcTxResult.TxID, actualAmount)
	}
	return &sdk.Result{
		Log:    logMsg,
		Data:   k.Codec().MustMarshalBinaryLengthPrefixed(isEqual),
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
	return btcutil.Amount(btcAmount.MulInt64(Bitcoin).RoundInt64()) == verifiedAmount
}
