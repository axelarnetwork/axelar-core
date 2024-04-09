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

func newTxWithUnwrappedMsgs(tx sdk.Tx) (txWithUnwrappedMsgs, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return txWithUnwrappedMsgs{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx must be a FeeTx")
	}

	return txWithUnwrappedMsgs{feeTx, unpackMsgs(tx.GetMsgs())}, nil
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
	tx, err := newTxWithUnwrappedMsgs(tx)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func unpackMsgs(msgs []sdk.Msg) []sdk.Msg {
	var unpackedMsgs []sdk.Msg

	for _, msg := range msgs {
		unpackedMsgs = append(unpackedMsgs, msg)

		if batchReq, ok := msg.(*auxiliarytypes.BatchRequest); ok {
			unpackedMsgs = append(unpackedMsgs, batchReq.UnwrapMessages()...)
		}
	}

	return unpackedMsgs
}
