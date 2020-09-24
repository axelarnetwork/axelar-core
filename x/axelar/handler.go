package axelar

import (
	"fmt"
	"github.com/axelarnetwork/axelar-core/x/axelar/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelar/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgTrackAddress:
			return handleMsgTrackAddress(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgTrackAddress(ctx sdk.Context, k keeper.Keeper, msg types.MsgTrackAddress) (*sdk.Result, error) {
	if err := k.TrackAddress(ctx, types.ExternalChainAddress{Chain: msg.Chain, Address: msg.Address}); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeTrackAddress),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeChain, msg.Chain),
			sdk.NewAttribute(types.AttributeAddress, string(msg.Address)),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
