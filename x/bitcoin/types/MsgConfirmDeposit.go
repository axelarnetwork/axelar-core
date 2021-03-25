package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgConfirmOutpoint represents a message to trigger the confirmation of a Bitcoin outpoint
type MsgConfirmOutpoint struct {
	Sender       sdk.AccAddress
	OutPointInfo OutPointInfo
}

// NewMsgConfirmOutpoint - MsgConfirmOutpoint constructor
func NewMsgConfirmOutpoint(sender sdk.AccAddress, out OutPointInfo) MsgConfirmOutpoint {
	return MsgConfirmOutpoint{
		Sender:       sender,
		OutPointInfo: out,
	}
}

// Route returns the route for this message
func (msg MsgConfirmOutpoint) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgConfirmOutpoint) Type() string {
	return "ConfirmOutpoint"
}

// ValidateBasic executes a stateless message validation
func (msg MsgConfirmOutpoint) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	if err := msg.OutPointInfo.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgConfirmOutpoint) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgConfirmOutpoint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
