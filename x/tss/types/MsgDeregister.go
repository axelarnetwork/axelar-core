package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgDeregister creates a message of type MsgDeregister
func NewMsgDeregister(sender sdk.AccAddress) *MsgDeregister {
	return &MsgDeregister{
		Sender: sender,
	}
}

// Route implements sdk.Msg
func (m MsgDeregister) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgDeregister) Type() string {
	return "Deregister"
}

// ValidateBasic implements sdk.Msg
func (m MsgDeregister) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m MsgDeregister) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m MsgDeregister) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
