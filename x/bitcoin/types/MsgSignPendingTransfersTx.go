package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgSignPendingTransfers struct {
	Sender sdk.AccAddress
	Fee    btcutil.Amount
}

func NewMsgSignPendingTransfers(sender sdk.AccAddress, fee btcutil.Amount) sdk.Msg {
	return MsgSignPendingTransfers{Sender: sender, Fee: fee}
}

func (msg MsgSignPendingTransfers) Route() string {
	return RouterKey
}

func (msg MsgSignPendingTransfers) Type() string {
	return "SignPendingTransfers"
}

func (msg MsgSignPendingTransfers) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Fee <= 0 {
		return fmt.Errorf("fee must be a positive amount")
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
