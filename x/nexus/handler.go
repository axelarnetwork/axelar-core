package nexus

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// NewHandler returns the handler of the nexus module
func NewHandler(k types.Nexus, snapshotter types.Snapshotter) sdk.Handler {
	server := keeper.NewMsgServerImpl(k, snapshotter)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.RegisterChainMaintainerRequest:
			res, err := server.RegisterChainMaintainer(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.DeregisterChainMaintainerRequest:
			res, err := server.DeregisterChainMaintainer(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		res, err := h(ctx, msg)
		if err != nil {
			k.Logger(ctx).Debug(err.Error())
			return nil, sdkerrors.Wrap(types.ErrNexus, err.Error())
		}

		if len(res.Log) > 0 {
			k.Logger(ctx).Debug(res.Log)
		}

		return res, nil
	}
}
