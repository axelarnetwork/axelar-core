package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgSignPendingTransfersTx struct {
	Sender sdk.AccAddress
}

func NewMsgSignPendingTransfersTx(sender sdk.AccAddress) sdk.Msg {
	return MsgSignPendingTransfersTx{Sender: sender}
}

func (msg MsgSignPendingTransfersTx) Route() string {
	return RouterKey
}

func (msg MsgSignPendingTransfersTx) Type() string {
	return "SignPendingTransfersTx"
}

func (msg MsgSignPendingTransfersTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

func (msg MsgSignPendingTransfersTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSignPendingTransfersTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
