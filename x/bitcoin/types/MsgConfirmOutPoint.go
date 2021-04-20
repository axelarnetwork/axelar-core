package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgConfirmOutpoint - MsgConfirmOutpoint constructor
func NewMsgConfirmOutpoint(sender sdk.AccAddress, out OutPointInfo) *MsgConfirmOutpoint {
	return &MsgConfirmOutpoint{
		Sender:       sender,
		OutPointInfo: out,
	}
}

// Route returns the route for this message
func (m MsgConfirmOutpoint) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m MsgConfirmOutpoint) Type() string {
	return "ConfirmOutpoint"
}

// ValidateBasic executes a stateless message validation
func (m MsgConfirmOutpoint) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.OutPointInfo.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgConfirmOutpoint) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgConfirmOutpoint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
