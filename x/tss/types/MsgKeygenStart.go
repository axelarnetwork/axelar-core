package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgKeygenStart indicate the start of keygen
type MsgKeygenStart struct {
	Sender     sdk.AccAddress
	NewKeyID   string
	SubsetSize int64
}

// NewMsgKeygenStart constructor for MsgKeygenStart
func NewMsgKeygenStart(sender sdk.AccAddress, newKeyID string, subsetSize int64) sdk.Msg {
	return MsgKeygenStart{
		Sender:     sender,
		NewKeyID:   newKeyID,
		SubsetSize: subsetSize,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgKeygenStart) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (msg MsgKeygenStart) Type() string { return "KeyGenStart" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgKeygenStart) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}

	if msg.NewKeyID == "" {
		return sdkerrors.Wrap(ErrTss, "new key id must be set")
	}

	if msg.SubsetSize < 0 {
		return sdkerrors.Wrap(ErrTss, "subset size has to be greater than or equal to 0")
	}

	// TODO enforce a maximum length for msg.NewKeyID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgKeygenStart) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

// GetSigners implements the sdk.Msg interface.
func (msg MsgKeygenStart) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
