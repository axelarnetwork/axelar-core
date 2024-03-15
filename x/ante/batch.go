package ante

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	batchertypes "github.com/axelarnetwork/axelar-core/x/batcher/types"
)

// messageWrapper implements the Tx interface for a slice of sdk messages
type messageWrapper struct {
	messages []sdk.Msg
}

func (m messageWrapper) ValidateBasic() error {
	for _, message := range m.messages {
		if err := message.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

func (m messageWrapper) GetMsgs() []sdk.Msg {
	return m.messages
}

func (m messageWrapper) Append(msg sdk.Msg) messageWrapper {
	m.messages = append(m.messages, msg)

	return m
}

// BatchDecorator runs anteHandler on the inner messages of a batch request
type BatchDecorator struct {
	cdc         codec.Codec
	anteHandler sdk.AnteHandler
}

// NewBatchDecorator is the constructor for BatchDecorator
func NewBatchDecorator(cdc codec.Codec, anteHandler sdk.AnteHandler) BatchDecorator {
	return BatchDecorator{
		cdc,
		anteHandler,
	}
}

// AnteHandle record qualified refund for the multiSig and vote transactions
func (b BatchDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		switch req := msg.(type) {
		case *batchertypes.BatchRequest:
			var messages messageWrapper
			var err error

			for _, m := range req.Messages {
				var sdkMsg sdk.Msg
				if err = b.cdc.UnpackAny(&m, &sdkMsg); err != nil {
					return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("unpack failed: %s", err))
				}

				if !msg.GetSigners()[0].Equals(req.Sender) {
					return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("message signer mismatch"))
				}

				messages = messages.Append(sdkMsg)
			}

			ctx, err = b.anteHandler(ctx, messages, simulate)
			if err != nil {
				return ctx, err
			}
		default:
			continue
		}

	}

	return next(ctx, tx, simulate)
}
