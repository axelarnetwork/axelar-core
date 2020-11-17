package voting

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/voting/keeper"
	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgBallot:
			return handleMsgBatchVote(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgBatchVote(ctx sdk.Context, k keeper.Keeper, msg types.MsgBallot) (*sdk.Result, error) {
	k.Logger(ctx).Debug("Handle batched votes")

	if err := k.ProcessBallot(ctx, msg.Sender, msg.Votes); err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeRecordVotes),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeAddress, fmt.Sprintf("%v", msg.Votes)),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
