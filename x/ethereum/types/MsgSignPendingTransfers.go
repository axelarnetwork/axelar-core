package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSignPendingTransfers represents a message to trigger the signing of all pending transfers
type MsgSignPendingTransfers struct {
	Sender sdk.AccAddress
}

// NewMsgSignPendingTransfers - MsgSignPendingTransfers constructor
func NewMsgSignPendingTransfers(sender sdk.AccAddress) sdk.Msg {
	return MsgSignPendingTransfers{Sender: sender}
}

// Route returns the route for this message
func (msg MsgSignPendingTransfers) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSignPendingTransfers) Type() string {
	return "SignPendingTransfers"
}

// ValidateBasic executes a stateless message validation
func (msg MsgSignPendingTransfers) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgSignPendingTransfers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgSignPendingTransfers) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
