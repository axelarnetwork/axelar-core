package tss

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTSS:
			return handleMsgTSS(ctx, k, &msg)
		case types.MsgKeygenStart:
			return handleMsgKeygenStart(ctx, &k, msg)
		case types.MsgSignStart:
			return handleMsgSignStart(ctx, &k, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgTSS(ctx sdk.Context, k keeper.Keeper, msg *types.MsgTSS) (*sdk.Result, error) {
	if err := k.KeygenMsg(ctx, msg); err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeMsgIn),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyPayload, msg.Payload.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgKeygenStart(ctx sdk.Context, k *keeper.Keeper, msg types.MsgKeygenStart) (*sdk.Result, error) {
	if err := k.StartKeygen(ctx, msg); err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeMsgKeygenStart),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			// sdk.NewAttribute(types.AttributePayload, msg.Payload.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgSignStart(ctx sdk.Context, k *keeper.Keeper, msg types.MsgSignStart) (*sdk.Result, error) {
	if err := k.StartSign(ctx, msg); err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeMsgSignStart),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			// sdk.NewAttribute(types.AttributePayload, msg.Payload.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
