package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/batch/types"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	cdc    codec.Codec
	router *baseapp.MsgServiceRouter
}

func NewMsgServer(cdc codec.Codec, msgServiceRouter *baseapp.MsgServiceRouter) types.MsgServiceServer {
	return msgServer{
		cdc,
		msgServiceRouter,
	}
}

func (s msgServer) Batch(c context.Context, req *types.BatchRequest) (*types.BatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var results []*sdk.Result
	var failedMessages []types.FailedMessages_FailedMessage

	for i, message := range req.UnwrapMessages() {
		cacheCtx, writeCache := ctx.CacheContext()

		res, err := s.processMessage(cacheCtx, message)
		if err != nil {
			failedMessages = append(failedMessages, types.FailedMessages_FailedMessage{
				Index: int32(i),
				Error: err.Error(),
			})
		} else {
			writeCache()
			ctx.EventManager().EmitEvents(res.GetEvents())
		}

		results = append(results, res)
	}

	if len(failedMessages) > 0 {
		events.Emit(ctx, &types.FailedMessages{
			Messages: failedMessages,
		})
	}

	return &types.BatchResponse{
		Results: results,
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
