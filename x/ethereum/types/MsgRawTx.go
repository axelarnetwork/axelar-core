package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type MsgRawTx struct {
	Sender sdk.AccAddress
	Tx     *ethTypes.Transaction
}

func NewMsgRawTx(sender sdk.AccAddress, tx *ethTypes.Transaction) sdk.Msg {
	return MsgRawTx{
		Sender: sender,
		Tx:     tx,
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
	if msg.Tx == nil {
		return fmt.Errorf("missing tx")
	}
	if msg.Tx.Data() == nil {
		return fmt.Errorf("missing smart contract call data")
	}
	if msg.Tx.To() == nil {
		return fmt.Errorf("missing recipient")
	}
	if msg.Tx.GasPrice() == nil {
		return fmt.Errorf("missing gas price")
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
