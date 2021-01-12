package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	Tx     *ethTypes.Transaction
}

func NewMsgVerifyTx(sender sdk.AccAddress, tx *ethTypes.Transaction) MsgVerifyTx {
	return MsgVerifyTx{
		Sender: sender,
		Tx:     tx,
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
	if msg.Tx.Data() == nil {
		return fmt.Errorf("missing smart contract call data")
	}
	if msg.Tx.Value() == nil {
		return fmt.Errorf("missing tx value")
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
