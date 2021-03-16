package snapshot

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgSnapshot:
			return handleMsgSnapshot(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgSnapshot(ctx sdk.Context, k keeper.Keeper, msg types.MsgSnapshot) (*sdk.Result, error) {
	if err := k.TakeSnapshot(ctx); err != nil {
		return nil, err
	}

	// if the snapshot was successful, we can be sure it will be retrieved
	snapshot, _ := k.GetLatestSnapshot(ctx, true)

	k.Logger(ctx).Info(
		fmt.Sprintf("Successfully obtained snapshot for counter %d with %d validators holding a total sum of %d voting power",
			k.GetLatestCounter(ctx), len(snapshot.Validators), snapshot.TotalPower.Int64()))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeModule),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
