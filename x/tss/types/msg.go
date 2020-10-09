package types

import (
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// golang stupidity: ensure interface compliance at compile time
var (
	_ sdk.Msg = &MsgKeygenStart{}
	_ sdk.Msg = &MsgIn{}
)

// MsgKeygenStart indicate the start of keygen
type MsgKeygenStart struct {
	Sender  sdk.AccAddress
	Payload *tssd.KeygenInfo
}

// MsgIn incoming message for either keygen or sign
// TODO it should be MsgOut! that's what's pushed to the chain; it's my job to convert it to a tssd.MessageIn for KeygenMsg
type MsgIn struct {
	Sender  sdk.AccAddress
	Payload *tssd.MessageIn
}

// NewMsgKeygenStart TODO unnecessary method; delete it?
func NewMsgKeygenStart(sender sdk.AccAddress, payload *tssd.KeygenInfo) MsgKeygenStart {
	return MsgKeygenStart{
		Sender:  sender,
		Payload: payload,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgKeygenStart) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msg.go
func (msg MsgKeygenStart) Type() string { return "keygen_start" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgKeygenStart) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender must be set")
	}
	// if msg.Chain == "" {
	// 	return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "name of the chain for address must be set")
	// }
	// if msg.Address == "" {
	// 	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address must be set")
	// }

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

// NewMsgIn TODO unnecessary method; delete it?
func NewMsgIn(sender sdk.AccAddress, payload *tssd.MessageIn) MsgIn {
	return MsgIn{
		Sender:  sender,
		Payload: payload,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgIn) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msg.go
func (msg MsgIn) Type() string { return "in" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgIn) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender must be set")
	}
	// if msg.Chain == "" {
	// 	return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "name of the chain for address must be set")
	// }
	// if msg.Address == "" {
	// 	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address must be set")
	// }

	return nil
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgIn) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements the sdk.Msg interface.
func (msg MsgIn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
