package ante

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	batchertypes "github.com/axelarnetwork/axelar-core/x/batcher/types"
)

// BatchDecorator implements the Tx interface and runs anteHandler on the inner messages of a batch request
type BatchDecorator struct {
	anteHandler sdk.AnteHandler
	cdc         codec.Codec
	messages    []sdk.Msg
}

// NewBatchDecorator is the constructor for BatchDecorator
func NewBatchDecorator(cdc codec.Codec, anteHandler sdk.AnteHandler) BatchDecorator {
	return BatchDecorator{
		anteHandler,
		cdc,
		[]sdk.Msg{},
	}
}

// AnteHandle record qualified refund for the multiSig and vote transactions
func (b BatchDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	for _, msg := range msgs {
		switch req := msg.(type) {
		case *batchertypes.BatchRequest:
			var err error

			for _, m := range req.Messages {
				var sdkMsg sdk.Msg
				if err = b.cdc.UnpackAny(&m, &sdkMsg); err != nil {
					return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("unpack failed: %s", err))
				}

				if !msg.GetSigners()[0].Equals(req.Sender) {
					return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("message signer mismatch"))
				}

				b.messages = append(b.messages, sdkMsg)
			}

			ctx, err = b.anteHandler(ctx, b, simulate)
			if err != nil {
				return ctx, err
			}
		default:
			continue
		}

	}

	return next(ctx, tx, simulate)
}

func (b BatchDecorator) ValidateBasic() error {
	for _, message := range b.messages {
		if err := message.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

func (b BatchDecorator) GetMsgs() []sdk.Msg {
	return b.messages
}
