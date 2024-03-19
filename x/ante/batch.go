package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	batchtypes "github.com/axelarnetwork/axelar-core/x/batch/types"
)

var _ sdk.FeeTx = (*BatchDecorator)(nil)

// BatchDecorator implements the Tx interface and runs anteHandler on the inner messages of a batch request
type BatchDecorator struct {
	sdk.FeeTx

	cdc      codec.Codec
	messages []sdk.Msg
}

// NewBatchDecorator is the constructor for BatchDecorator
func NewBatchDecorator(cdc codec.Codec) BatchDecorator {
	return BatchDecorator{
		nil,
		cdc,
		[]sdk.Msg{},
	}
}

// AnteHandle record qualified refund for the multiSig and vote transactions
func (b BatchDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	if !anyBatch(msgs) {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx must be a FeeTx")
	}

	b.FeeTx = feeTx

	for _, msg := range msgs {
		switch req := msg.(type) {
		case *batchtypes.BatchRequest:
			innerMsgs := req.UnwrapMessages()
			if anyBatch(innerMsgs) {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "nested batch requests are not allowed")
			}

			b.messages = append(b.messages, innerMsgs...)

		default:
			b.messages = append(b.messages, msg)
		}

	}

	return next(ctx, b, simulate)
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

func anyBatch(msgs []sdk.Msg) bool {
	if len(msgs) == 0 {
		return false
	}

	for _, msg := range msgs {
		switch msg.(type) {
		case *batchtypes.BatchRequest:
			return true
		}
	}
	return false
}
