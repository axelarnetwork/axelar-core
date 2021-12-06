package permission

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/permission/keeper"
	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

// NewHandler returns the handler of the Cosmos module
func NewHandler(k keeper.Keeper) sdk.Handler {
	server := keeper.NewMsgServerImpl(k)
	h := func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.UpdateGovernanceKeyRequest:
			res, err := server.UpdateGovernanceKey(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)

			return result, err
		case *types.RegisterControllerRequest:
			res, err := server.RegisterController(sdk.WrapSDKContext(ctx), msg)
			result, err := sdk.WrapServiceResult(ctx, res, err)

			return result, err
		case *types.DeregisterControllerRequest:
			res, err := server.DeregisterController(sdk.WrapSDKContext(ctx), msg)
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
			return nil, sdkerrors.Wrap(types.ErrGov, err.Error())
		}
		return res, nil
	}
}
