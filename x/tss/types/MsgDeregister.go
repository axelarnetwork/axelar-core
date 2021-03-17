package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgDeregister to deregister so that the validator will not participate in any future keygen
type MsgDeregister struct {
	Sender sdk.AccAddress
}

// NewMsgDeregister creates a message of type MsgDeregister
func NewMsgDeregister(sender sdk.AccAddress) sdk.Msg {
	return MsgDeregister{
		Sender: sender,
	}
}

// Route implements sdk.Msg
func (msg MsgDeregister) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgDeregister) Type() string {
	return "Deregister"
}

// ValidateBasic implements sdk.Msg
func (msg MsgDeregister) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgDeregister) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgDeregister) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
