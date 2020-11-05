package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgWithdraw{}

type MsgWithdraw struct {
	Sender      sdk.AccAddress
	TxID        string
	SignatureID string
}

func NewMsgWithdraw(sender sdk.AccAddress, txId string, sigId string) MsgWithdraw {
	return MsgWithdraw{
		Sender:      sender,
		TxID:        txId,
		SignatureID: sigId,
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
		return fmt.Errorf("invalid tx ID: %s", msg.TxID)
	}
	if msg.SignatureID == "" {
		return fmt.Errorf("invalid signature ID: %s", msg.SignatureID)
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
