package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSignBurnTokens represents the message to sign commands to burn tokens with AxelarGateway
type MsgSignBurnTokens struct {
	Sender sdk.AccAddress
}

// NewMsgSignBurnTokens is the constructor for MsgSignBurnTokens
func NewMsgSignBurnTokens(sender sdk.AccAddress) sdk.Msg {
	return MsgSignBurnTokens{Sender: sender}
}

// Route implements sdk.Msg
func (msg MsgSignBurnTokens) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgSignBurnTokens) Type() string {
	return "SignBurnTokens"
}

// GetSignBytes  implements sdk.Msg
func (msg MsgSignBurnTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgSignBurnTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// ValidateBasic implements sdk.Msg
func (msg MsgSignBurnTokens) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}
