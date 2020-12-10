package btc_bridge

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/eth_bridge/keeper"
	"github.com/axelarnetwork/axelar-core/x/eth_bridge/types"
)

func NewHandler(k keeper.Keeper, rpc types.RPCClient, v types.Voter, s types.Signer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgWithdraw:
			return handleMsgWithdraw(ctx, k, rpc, s, msg)

		case types.MsgVerifyTx:
			return handleMsgVerifyTx(ctx, k, rpc, v, msg)

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest,
				fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
	}
}

func handleMsgWithdraw(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, signer types.Signer, msg types.MsgWithdraw) (*sdk.Result, error) {
	k.Logger(ctx).Debug(fmt.Sprintf("start tracking address"))

	panic("implement me")
}

func handleMsgVerifyTx(ctx sdk.Context, k keeper.Keeper, rpc types.RPCClient, v types.Voter, msg types.MsgVerifyTx) (*sdk.Result, error) {
	k.Logger(ctx).Debug("verifying bitcoin transaction")

	panic("implement me")
}
