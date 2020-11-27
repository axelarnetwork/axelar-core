package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgWithdraw struct {
	Sender       sdk.AccAddress
	TxID         string
	SignatureID  string
	KeyID        string
	UseMasterKey bool
}

func NewMsgWithdrawWithMasterKey(sender sdk.AccAddress, txId string, sigId string) sdk.Msg {
	return MsgWithdraw{
		Sender:       sender,
		TxID:         txId,
		SignatureID:  sigId,
		UseMasterKey: true,
	}
}

func NewMsgWithdraw(sender sdk.AccAddress, txId string, sigId string, keyId string) sdk.Msg {
	return MsgWithdraw{
		Sender:      sender,
		TxID:        txId,
		SignatureID: sigId,
		KeyID:       keyId,
	}
}

func (msg MsgWithdraw) Route() string {
	return RouterKey
}

func (msg MsgWithdraw) Type() string {
	return "Withdraw"
}

func (msg MsgWithdraw) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.TxID == "" {
		return fmt.Errorf("missing transaction ID")
	}
	if msg.SignatureID == "" {
		return fmt.Errorf("missing signature ID")
	}
	if msg.KeyID == "" && !msg.UseMasterKey {
		return fmt.Errorf("missing key ID")
	}

	return nil
}

func (msg MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
