package types

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSignTx  is a message struct to make a new raw transaction known to the blockchain
type MsgSignTx struct {
	Sender sdk.AccAddress
	TxID   string
	RawTx  *wire.MsgTx
}

// NewMsgSignTx creates a new MsgSignTx
func NewMsgSignTx(sender sdk.AccAddress, txID string, rawTx *wire.MsgTx) MsgSignTx {
	return MsgSignTx{
		Sender: sender,
		TxID:   txID,
		RawTx:  rawTx,
	}
}

// Route returns the module route
func (msg MsgSignTx) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSignTx) Type() string {
	return "RawTx"
}

// ValidateBasic performs a stateless check of the message content
func (msg MsgSignTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.TxID == "" {
		return fmt.Errorf("missing txID")
	}
	if msg.RawTx.TxIn == nil {
		return fmt.Errorf("missing txIn")
	}
	if msg.RawTx.TxOut == nil {
		return fmt.Errorf("missing txOut")
	}
	if len(msg.RawTx.TxOut) != 1 {
		return fmt.Errorf("expected exactly 1 txOut, got %d", len(msg.RawTx.TxOut))
	}

	return nil
}

// GetSignBytes returns the serialization of the message
func (msg MsgSignTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the signer of the message
func (msg MsgSignTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
