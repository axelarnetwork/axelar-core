package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgSendTx struct {
	Sender      sdk.AccAddress
	TxID        string
	SignatureID string
}

func NewMsgSendTx(sender sdk.AccAddress, txId string, sigId string) MsgSendTx {
	return MsgSendTx{
		Sender:      sender,
		TxID:        txId,
		SignatureID: sigId,
	}
}

func (msg MsgSendTx) Route() string {
	return RouterKey
}

func (msg MsgSendTx) Type() string {
	return "Send"
}

func (msg MsgSendTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.TxID == "" {
		return fmt.Errorf("missing transaction ID")
	}
	if msg.SignatureID == "" {
		return fmt.Errorf("missing signature ID")
	}

	return nil
}

func (msg MsgSendTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSendTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
