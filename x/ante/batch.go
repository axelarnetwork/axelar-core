package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	auxiliarytypes "github.com/axelarnetwork/axelar-core/x/auxiliary/types"
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
	unwrappedMsgs, err := unpackMsgs(tx.GetMsgs())
	if err != nil {
		return ctx, err
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx must be a FeeTx")
	}

	return next(ctx, txWithUnwrappedMsgs{feeTx, unwrappedMsgs}, simulate)
}

func unpackMsgs(msgs []sdk.Msg) ([]sdk.Msg, error) {
	var unpackedMsgs []sdk.Msg
	idx := 0

	for i, msg := range msgs {
		if batchReq, ok := msg.(*auxiliarytypes.BatchRequest); ok {
			// Bulk append messages, including the current batch request
			unpackedMsgs = append(unpackedMsgs, msgs[idx:i+1]...)

			// Unwrap the batch request and append the messages
			unpackedMsgs = append(unpackedMsgs, batchReq.UnwrapMessages()...)

			idx = i + 1
		}
	}

	// avoid copying the slice if there are no batch requests
	if len(unpackedMsgs) == 0 {
		return msgs, nil
	}

	return append(unpackedMsgs, msgs[idx:]...), nil
}
