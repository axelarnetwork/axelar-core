package tss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// NewHandler returns the handler for the tss module
func NewHandler(k keeper.Keeper, s types.Snapshotter, n types.Nexus, v types.Voter, staker types.StakingKeeper, rewarder types.Rewarder) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, s, staker, v, n, rewarder)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.HeartBeatRequest:
			res, err := server.HeartBeat(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)
			return result, err
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			k.Logger(ctx).Debug(err.Error())
			return nil, sdkerrors.Wrap(types.ErrTss, err.Error())
		}
		if res.Log != "" {
			k.Logger(ctx).Debug(res.Log)
		}
		return res, nil
	}
}
