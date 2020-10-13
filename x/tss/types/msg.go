package types

import (
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// golang stupidity: ensure interface compliance at compile time
var (
	_ sdk.Msg = &MsgKeygenStart{}
	_ sdk.Msg = &MsgTSS{}
)

// MsgKeygenStart indicate the start of keygen
type MsgKeygenStart struct {
	// Sender  sdk.AccAddress
	// Payload *tssd.KeygenInfo
	NewKeyID  string
	Threshold int
}

// MsgTSS protocol message for either keygen or sign
type MsgTSS struct {
	Sender    sdk.AccAddress
	SessionID string
	Payload   *tssd.MessageOut // TODO probably should not be a pointer; it's serialized by cosmos
}

// NewMsgKeygenStart TODO unnecessary method; delete it?
func NewMsgKeygenStart(newKeyID string, threshold int) MsgKeygenStart {
	return MsgKeygenStart{
		NewKeyID:  newKeyID,
		Threshold: threshold,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgKeygenStart) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msg.go
func (msg MsgKeygenStart) Type() string { return "keygen_start" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgKeygenStart) ValidateBasic() error {
	// if msg.Sender == nil {
	// 	return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender must be set")
	// }
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
	// return []sdk.AccAddress{msg.Sender}
	return []sdk.AccAddress{}
}

// NewMsgTSS TODO unnecessary method; delete it?
func NewMsgTSS(sender sdk.AccAddress, sessionID string, payload *tssd.MessageOut) MsgTSS {
	return MsgTSS{
		Sender:  sender,
		Payload: payload,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgTSS) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msg.go
func (msg MsgTSS) Type() string { return "in" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgTSS) ValidateBasic() error {
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
func (msg MsgTSS) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements the sdk.Msg interface.
func (msg MsgTSS) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
