package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgTransferToNewMasterKey struct {
	Sender      sdk.AccAddress
	TxID        string
	SignatureID string
}

func NewMsgTransferToNewMasterKey(sender sdk.AccAddress, txId string, sigId string) sdk.Msg {
	return MsgTransferToNewMasterKey{
		Sender:      sender,
		TxID:        txId,
		SignatureID: sigId,
	}
}
func (msg MsgTransferToNewMasterKey) Route() string {
	return RouterKey
}

func (msg MsgTransferToNewMasterKey) Type() string {
	return "MsgTransferToNewMasterKey"
}

func (msg MsgTransferToNewMasterKey) ValidateBasic() error {
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

func (msg MsgTransferToNewMasterKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTransferToNewMasterKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
