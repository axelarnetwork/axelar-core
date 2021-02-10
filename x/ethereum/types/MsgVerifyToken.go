package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgVerifyToken struct {
	Sender       sdk.AccAddress
	ContractAddr string
	Symbol       string
}

func NewMsgVerifyToken(sender sdk.AccAddress, contract, symbol string) sdk.Msg {
	return MsgVerifyToken{
		Sender:       sender,
		ContractAddr: contract,
		Symbol:       symbol,
	}
}

func (msg MsgVerifyToken) Route() string {
	return RouterKey
}

func (msg MsgVerifyToken) Type() string {
	return "VerifyToken"
}

func (msg MsgVerifyToken) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

func (msg MsgVerifyToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVerifyToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
