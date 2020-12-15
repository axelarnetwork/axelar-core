package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	TX     TX
}

func NewMsgVerifyTx(sender sdk.AccAddress, hash *common.Hash, address EthAddress, amount big.Int, txType TXType) sdk.Msg {
	return MsgVerifyTx{
		Sender: sender,
		TX: TX{
			Hash:    hash,
			Address: address,
			Amount:  amount,
			TXType:  txType,
		},
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

	return nil
}

func (msg MsgVerifyTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVerifyTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
