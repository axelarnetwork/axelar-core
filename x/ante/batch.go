package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	batchtypes "github.com/axelarnetwork/axelar-core/x/batch/types"
)

// txWithUnwrappedMsgs implements the FeeTx interface
type txWithUnwrappedMsgs struct {
	sdk.FeeTx
	messages []sdk.Msg
}

func (t txWithUnwrappedMsgs) ValidateBasic() error {
	for _, message := range t.messages {
		if err := message.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

func (t txWithUnwrappedMsgs) GetMsgs() []sdk.Msg {
	return t.messages
}

// BatchDecorator unwraps batch requests and passes them to the next AnteHandler
type BatchDecorator struct {
	cdc codec.Codec
}

// NewBatchDecorator is the constructor for BatchDecorator
func NewBatchDecorator(cdc codec.Codec) BatchDecorator {
	return BatchDecorator{
		cdc,
	}
}

// AnteHandle record qualified refund for the multiSig and vote transactions
func (b BatchDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()

	if !batchtypes.AnyBatch(msgs) {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx must be a FeeTx")
	}

	var unwrappedMsgs []sdk.Msg
	start := 0
	for i, msg := range msgs {
		if batchReq, ok := msg.(*batchtypes.BatchRequest); ok {
			// Bulk append messages, including the current batch request
			unwrappedMsgs = append(unwrappedMsgs, msgs[start:i+1]...)

			innerMsgs := batchReq.UnwrapMessages()
			if batchtypes.AnyBatch(innerMsgs) {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "nested batch requests are not allowed")
			}
			unwrappedMsgs = append(unwrappedMsgs, innerMsgs...)

			start = i + 1
		}
	}

	if len(unwrappedMsgs) == 0 {
		return next(ctx, tx, simulate)
	}

	return next(ctx, txWithUnwrappedMsgs{feeTx, unwrappedMsgs}, simulate)
}
