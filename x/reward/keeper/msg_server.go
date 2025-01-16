package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/axelarnetwork/axelar-core/x/reward/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	types.Refunder
	msgSvcRouter *baseapp.MsgServiceRouter
	bank         types.Banker
}

// NewMsgServerImpl returns a new msg server instance
func NewMsgServerImpl(k types.Refunder, b types.Banker, m *baseapp.MsgServiceRouter) types.MsgServiceServer {
	return msgServer{
		Refunder:     k,
		bank:         b,
		msgSvcRouter: m,
	}
}

// RefundMsg refunds the fees of the inner message upon correct execution
func (s msgServer) RefundMsg(c context.Context, req *types.RefundMsgRequest) (*types.RefundMsgResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	msg := req.GetInnerMessage()
	if msg == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid inner message")
	}

	result, err := s.routeInnerMsg(ctx, msg)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to execute message")
	}

	refund, found := s.Refunder.GetPendingRefund(ctx, *req)
	if found {
		// refund tx fee to the given account.
		err = s.bank.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, refund.Payer, refund.Fees)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to refund tx fee")
		}

		s.Refunder.DeletePendingRefund(ctx, *req)
	}

	ctx.EventManager().EmitEvents(result.GetEvents())

	return &types.RefundMsgResponse{Data: result.Data, Log: result.Log}, nil
}

func (s msgServer) routeInnerMsg(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
	var msgResult *sdk.Result
	var err error

	if handler := s.msgSvcRouter.Handler(msg); handler != nil {
		// ADR 031 request type routing
		msgResult, err = handler(ctx, msg)
	} else {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
	}

	return msgResult, err
}
