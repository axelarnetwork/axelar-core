package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = MsgSnapshot{}

type MsgSnapshot struct {
	Sender sdk.AccAddress
}

func (msg MsgSnapshot) Route() string {
	return RouterKey
}

func (msg MsgSnapshot) Type() string {
	return "Snapshot"
}

func (msg MsgSnapshot) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender must be set")
	}

	return nil
}

func (msg MsgSnapshot) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSnapshot) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
