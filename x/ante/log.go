package ante

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
)

// LogMsgDecorator logs all messages in blocks
type LogMsgDecorator struct {
	cdc codec.Codec
}

// NewLogMsgDecorator is the constructor for LogMsgDecorator
func NewLogMsgDecorator(cdc codec.Codec) LogMsgDecorator {
	return LogMsgDecorator{cdc: cdc}
}

// AnteHandle logs all messages in blocks
func (d LogMsgDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate || ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		logger(ctx).Debug(fmt.Sprintf("received message of type %s in block %d: %s",
			proto.MessageName(msg),
			ctx.BlockHeight(),
			string(d.cdc.MustMarshalJSON(msg)),
		))
	}

	return next(ctx, tx, simulate)
}
