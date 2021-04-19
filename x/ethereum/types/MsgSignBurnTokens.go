package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgSignBurnTokens is the constructor for MsgSignBurnTokens
func NewMsgSignBurnTokens(sender sdk.AccAddress) *MsgSignBurnTokens {
	return &MsgSignBurnTokens{Sender: sender}
}

// Route implements sdk.Msg
func (m MsgSignBurnTokens) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgSignBurnTokens) Type() string {
	return "SignBurnTokens"
}

// GetSignBytes  implements sdk.Msg
func (m MsgSignBurnTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m MsgSignBurnTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m MsgSignBurnTokens) ValidateBasic() error {
	if m.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}
