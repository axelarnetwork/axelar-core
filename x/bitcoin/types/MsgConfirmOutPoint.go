package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgConfirmOutpoint - MsgConfirmOutpoint constructor
func NewMsgConfirmOutpoint(sender sdk.AccAddress, out OutPointInfo) *MsgConfirmOutpoint {
	return &MsgConfirmOutpoint{
		Sender:       sender.String(),
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
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
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
	from, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
