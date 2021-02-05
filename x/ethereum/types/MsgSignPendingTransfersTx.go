package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgSignPendingTransfers struct {
	Sender sdk.AccAddress
}

func NewMsgSignPendingTransfersTx(sender sdk.AccAddress) sdk.Msg {
	return MsgSignPendingTransfers{Sender: sender}
}

func (msg MsgSignPendingTransfers) Route() string {
	return RouterKey
}

func (msg MsgSignPendingTransfers) Type() string {
	return "SignPendingTransfersTx"
}

func (msg MsgSignPendingTransfers) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

func (msg MsgSignPendingTransfers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSignPendingTransfers) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
