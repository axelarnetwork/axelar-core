package broadcast

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

// NewHandler returns a new handler
func NewHandler(b exported.Broadcaster) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgRegisterProxy:
			return handleMsgRegisterProxy(ctx, b, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgRegisterProxy(ctx sdk.Context, b exported.Broadcaster, msg *types.MsgRegisterProxy) (*sdk.Result, error) {
	if err := b.RegisterProxy(ctx, msg.PrincipalAddr, msg.ProxyAddr); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBroadcast, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.PrincipalAddr.String()),
			sdk.NewAttribute(types.AttributeAddress, msg.ProxyAddr.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
