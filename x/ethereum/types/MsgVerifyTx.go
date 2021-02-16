package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	TxInfo TransactionInfo
}

func NewMsgVerifyTx(sender sdk.AccAddress, info TransactionInfo) MsgVerifyTx {
	return MsgVerifyTx{
		Sender: sender,
		TxInfo: info,
	}
}

func (msg MsgVerifyTx) Route() string {
	return RouterKey
}

func (msg MsgVerifyTx) Type() string {
	return "VerifyTx"
}

func (msg MsgVerifyTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

func (msg MsgVerifyTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVerifyTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
