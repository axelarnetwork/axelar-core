package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	Tx     []byte
}

func NewMsgVerifyTx(sender sdk.AccAddress, json []byte) MsgVerifyTx {
	return MsgVerifyTx{
		Sender: sender,
		Tx:     json,
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
	if msg.Tx == nil {
		return fmt.Errorf("missing tx")
	}
	tx := &ethTypes.Transaction{}
	err := tx.UnmarshalJSON(msg.Tx)
	if err != nil {
		return err
	}
	if tx.Data() == nil {
		return fmt.Errorf("missing smart contract call data")
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

func (msg MsgVerifyTx) UnmarshaledTx() *ethTypes.Transaction {
	tx := &ethTypes.Transaction{}
	err := tx.UnmarshalJSON(msg.Tx)
	if err != nil {
		panic(err)
	}
	return tx
}
