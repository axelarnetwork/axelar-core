package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgRawTx{}

type MsgRawTx struct {
	Sender      sdk.AccAddress
	TxID        string
	Amount      btcutil.Amount
	Destination string
}

func NewMsgRawTx(sender sdk.AccAddress, txId string, amount btcutil.Amount, destination string) MsgRawTx {
	return MsgRawTx{
		Sender:      sender,
		TxID:        txId,
		Amount:      amount,
		Destination: destination,
	}
}

func (msg MsgRawTx) Route() string {
	return RouterKey
}

func (msg MsgRawTx) Type() string {
	return "RawTx"
}

func (msg MsgRawTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.TxID == "" {
		return fmt.Errorf("invalid tx ID: %s", msg.TxID)
	}
	if msg.Amount <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero")
	}
	if msg.Destination == "" {
		return fmt.Errorf("missing destination")
	}

	return nil
}

func (msg MsgRawTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRawTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
