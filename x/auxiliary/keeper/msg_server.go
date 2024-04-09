package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	router *baseapp.MsgServiceRouter
}

func NewMsgServer(msgServiceRouter *baseapp.MsgServiceRouter) types.MsgServiceServer {
	return msgServer{
		msgServiceRouter,
	}
}

func (s msgServer) Batch(c context.Context, req *types.BatchRequest) (*types.BatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var results []types.BatchResponse_Response

	for i, message := range req.UnwrapMessages() {
		var batchResponse types.BatchResponse_Response

		cacheCtx, writeCache := ctx.CacheContext()
		res, err := s.processMessage(cacheCtx, message)
		if err != nil {
			batchResponse = types.BatchResponse_Response{Res: &types.BatchResponse_Response_Err{Err: err.Error()}}

			events.Emit(ctx, &types.BatchedMessageFailed{
				Index: int32(i),
				Error: err.Error(),
			})
		} else {
			batchResponse = types.BatchResponse_Response{Res: &types.BatchResponse_Response_Result{Result: res}}

			writeCache()
			ctx.EventManager().EmitEvents(res.GetEvents())
		}

		results = append(results, batchResponse)
	}

	return &types.BatchResponse{
		Responses: results,
	}, nil
}

func (s msgServer) processMessage(ctx sdk.Context, message sdk.Msg) (*sdk.Result, error) {
	handler := s.router.Handler(message)
	if handler == nil {
		return nil, fmt.Errorf("unrecognized message type: %s", sdk.MsgTypeURL(message))
	}

	res, err := handler(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %s", err)
	}

	return res, nil
}
