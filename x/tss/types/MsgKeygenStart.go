package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgKeygenStart constructor for MsgKeygenStart
func NewMsgKeygenStart(sender sdk.AccAddress, newKeyID string, subsetSize int64) *MsgKeygenStart {
	return &MsgKeygenStart{
		Sender:     sender,
		NewKeyID:   newKeyID,
		SubsetSize: subsetSize,
	}
}

// Route implements the sdk.Msg interface.
func (m MsgKeygenStart) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m MsgKeygenStart) Type() string { return "KeyGenStart" }

// ValidateBasic implements the sdk.Msg interface.
func (m MsgKeygenStart) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.NewKeyID == "" {
		return sdkerrors.Wrap(ErrTss, "new key id must be set")
	}

	if m.SubsetSize < 0 {
		return sdkerrors.Wrap(ErrTss, "subset size has to be greater than or equal to 0")
	}

	// TODO enforce a maximum length for m.NewKeyID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgKeygenStart) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m MsgKeygenStart) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
