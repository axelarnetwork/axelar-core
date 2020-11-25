package tss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func NewHandler(k keeper.Keeper, s types.Staker) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgKeygenTraffic:
			return handleMsgKeygenTraffic(ctx, k, msg)
		case types.MsgSignTraffic:
			return handleMsgSignTraffic(ctx, k, msg)
		case types.MsgKeygenStart:
			return handleMsgKeygenStart(ctx, k, s, msg)
		case types.MsgSignStart:
			return handleMsgSignStart(ctx, k, msg)
		case types.MsgMasterKeyRefresh:
			return handleMsgMasterKeyRefresh(ctx, k, s, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgMasterKeyRefresh(ctx sdk.Context, k keeper.Keeper, s types.Staker, msg types.MsgMasterKeyRefresh) (*sdk.Result, error) {
	snapshot, err := s.GetLatestSnapshot(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "key refresh failed")
	}
	if k.IsKeyRefreshLocked(ctx, snapshot.Timestamp) {
		return nil, fmt.Errorf("key refresh locked")
	}

	if err := k.StartKeyRefresh(ctx, snapshot.Validators); err != nil {
		return nil, sdkerrors.Wrap(err, "key refresh failed")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgKeygenTraffic(ctx sdk.Context, k keeper.Keeper, msg types.MsgKeygenTraffic) (*sdk.Result, error) {
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

func handleMsgKeygenStart(ctx sdk.Context, k keeper.Keeper, s types.Staker, msg types.MsgKeygenStart) (*sdk.Result, error) {
	snapshot, err := s.GetLatestSnapshot(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "key refresh failed")
	}

	if msg.Threshold < 1 || msg.Threshold > len(snapshot.Validators) {
		err := fmt.Errorf("invalid threshold: %d, validators: %d", msg.Threshold, len(snapshot.Validators))
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	if err := k.StartKeygen(ctx, msg.NewKeyID, msg.Threshold, snapshot.Validators); err != nil {
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

func handleMsgSignStart(ctx sdk.Context, k keeper.Keeper, msg types.MsgSignStart) (*sdk.Result, error) {
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

func handleMsgSignTraffic(ctx sdk.Context, k keeper.Keeper, msg types.MsgSignTraffic) (*sdk.Result, error) {
	if err := k.SignMsg(ctx, msg); err != nil {
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
