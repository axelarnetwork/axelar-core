package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/batcher/types"
	"github.com/axelarnetwork/utils/funcs"
)

var _ types.MsgServiceServer = msgServer{}

type msgServer struct {
	cdc        codec.Codec
	router     *baseapp.MsgServiceRouter
	anteHandle ante.MessageAnteHandler
}

func NewMsgServer(cdc codec.Codec, msgServiceRouter *baseapp.MsgServiceRouter, anteHandle ante.MessageAnteHandler) types.MsgServiceServer {
	return msgServer{
		cdc,
		msgServiceRouter,
		anteHandle,
	}
}

func (s msgServer) Batch(c context.Context, req *types.BatchRequest) (*types.BatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	processMessage := func(ctx sdk.Context, i int, message *cdctypes.Any) (*sdk.Result, error) {
		sdkMsg, err := unpackInnerMessage(s.cdc, message)
		if err != nil {
			return nil, fmt.Errorf("unpack failed at index %d: %s", i, err)
		}

		if err = req.ValidateInnerMessage(sdkMsg); err != nil {
			return nil, fmt.Errorf("message validation failed at index %d: %s", i, err)
		}

		handler := s.router.Handler(sdkMsg)
		if handler == nil {
			return nil, fmt.Errorf("unrecognized message type at index %d: %s", i, sdkMsg)
		}

		ctx, err = s.anteHandle(ctx, []sdk.Msg{sdkMsg}, false)
		if err != nil {
			return nil, fmt.Errorf("antehandler failed for message at index %d: %s", i, err)
		}

		res, err := handler(ctx, sdkMsg)
		if err != nil {
			return nil, fmt.Errorf("execution failed for message at index %d: %s", i, err)
		}

		return res, nil
	}

	for i, message := range req.MustSucceedMessages {
		if _, err := processMessage(ctx, i, &message); err != nil {
			return nil, err
		}
	}

	var failedMessages []types.FailedMessages_FailedMessage
	for i, message := range req.CanFailMessages {
		cacheCtx, writeCache := ctx.CacheContext()
		res, err := processMessage(cacheCtx, i, &message)
		if err != nil {
			failedMessages = append(failedMessages, types.FailedMessages_FailedMessage{
				Index: int32(i),
				Error: err.Error(),
			})
		} else {
			writeCache()
			ctx.EventManager().EmitEvents(res.GetEvents())
		}
	}

	if len(failedMessages) > 0 {
		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.FailedMessages{
			Messages: failedMessages,
		}))
	}

	return &types.BatchResponse{}, nil
}

func unpackInnerMessage(cdc codec.Codec, any *cdctypes.Any) (sdk.Msg, error) {
	var sdkMsg sdk.Msg
	if err := cdc.UnpackAny(any, &sdkMsg); err != nil {
		return nil, err
	}
	return sdkMsg, nil
}
