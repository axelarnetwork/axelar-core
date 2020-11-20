package voting

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/voting/keeper"
	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

func NewHandler(k keeper.Keeper, r sdk.Router) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgBallot:
			return handleMsgBallot(ctx, k, r, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgBallot(ctx sdk.Context, k keeper.Keeper, r sdk.Router, msg types.MsgBallot) (*sdk.Result, error) {
	k.Logger(ctx).Debug("Handle ballot")

	// routing the contained votes to the appropriate modules, setting the sender of each vote to the sender of the ballot
	for _, vote := range msg.Votes {
		vote.SetSender(msg.Sender)
		handler := r.Route(ctx, vote.Route())
		_, err := handler(ctx, vote)
		if err != nil {
			k.Logger(ctx).Info(sdkerrors.Wrap(err, fmt.Sprintf("vote for poll %s is invalid", vote.Poll().String())).Error())
		}

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
