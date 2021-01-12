package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type MsgSignTx struct {
	Sender sdk.AccAddress
	Tx     *ethTypes.Transaction
}

func NewMsgRawTx(sender sdk.AccAddress, tx *ethTypes.Transaction) sdk.Msg {
	return MsgSignTx{
		Sender: sender,
		Tx:     tx,
	}
}

func (msg MsgSignTx) Route() string {
	return RouterKey
}

func (msg MsgSignTx) Type() string {
	return "RawTx"
}

func (msg MsgSignTx) ValidateBasic() error {
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

func (msg MsgSignTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSignTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
