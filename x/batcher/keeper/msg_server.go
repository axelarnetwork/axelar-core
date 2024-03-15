package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/batcher/types"
	"github.com/axelarnetwork/utils/funcs"
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

	for i, message := range req.Messages {
		cacheCtx, writeCache := ctx.CacheContext()

		res, err := s.processMessage(cacheCtx, &message)
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
		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.FailedMessages{
			Messages: failedMessages,
		}))
	}

	return &types.BatchResponse{
		Results: results,
	}, nil
}

func (s msgServer) processMessage(ctx sdk.Context, message *cdctypes.Any) (*sdk.Result, error) {
	sdkMsg, err := unpackInnerMessage(s.cdc, message)
	if err != nil {
		return nil, fmt.Errorf("unpack failed: %s", err)
	}

	handler := s.router.Handler(sdkMsg)
	if handler == nil {
		return nil, fmt.Errorf("unrecognized message type: %s", sdk.MsgTypeURL(sdkMsg))
	}

	res, err := handler(ctx, sdkMsg)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %s", err)
	}

	return res, nil
}

func unpackInnerMessage(cdc codec.Codec, any *cdctypes.Any) (sdk.Msg, error) {
	var sdkMsg sdk.Msg
	if err := cdc.UnpackAny(any, &sdkMsg); err != nil {
		return nil, err
	}
	return sdkMsg, nil
}
