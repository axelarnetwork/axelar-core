package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type MsgSignTx struct {
	Sender sdk.AccAddress
	Tx     []byte
}

func NewMsgSignTx(sender sdk.AccAddress, jsonTx []byte) sdk.Msg {
	return MsgSignTx{
		Sender: sender,
		Tx:     jsonTx,
	}
}

func (msg MsgSignTx) Route() string {
	return RouterKey
}

func (msg MsgSignTx) Type() string {
	return "SignTx"
}

func (msg MsgSignTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Tx == nil {
		return fmt.Errorf("missing tx")
	}
	tx := ethTypes.Transaction{}
	if err := tx.UnmarshalJSON(msg.Tx); err != nil {
		return err
	}
	return nil
}

func (msg MsgSignTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSignTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg MsgSignTx) UnmarshaledTx() *ethTypes.Transaction {
	tx := &ethTypes.Transaction{}
	err := tx.UnmarshalJSON(msg.Tx)
	if err != nil {
		panic(err)
	}
	return tx
}
