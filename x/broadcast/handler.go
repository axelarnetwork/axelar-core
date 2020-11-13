package broadcast

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

func NewHandler(b exported.Broadcaster) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgRegisterProxy:
			return handleMsgRegisterProxy(ctx, b, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgRegisterProxy(ctx sdk.Context, b exported.Broadcaster, msg types.MsgRegisterProxy) (*sdk.Result, error) {
	if err := b.RegisterProxy(ctx, msg.Principal, msg.Proxy); err != nil {
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
