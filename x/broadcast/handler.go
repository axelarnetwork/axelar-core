package broadcast

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/broadcast/keeper"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgRegisterProxy:
			return handleMsgRegisterProxy(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgRegisterProxy(ctx sdk.Context, k keeper.Keeper, msg types.MsgRegisterProxy) (*sdk.Result, error) {
	if err := k.RegisterProxy(ctx, msg.Principal, msg.Proxy); err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeRegisterProxy),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Principal.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.Proxy.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
