package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgInstallSC struct {
	Sender          sdk.AccAddress
	SmartContractID string
	Bytecode        []byte
}

func NewMsgInstallSC(sender sdk.AccAddress, id string, bytecode []byte) sdk.Msg {
	return MsgInstallSC{
		Sender:          sender,
		SmartContractID: id,
		Bytecode:        bytecode,
	}

}

func (msg MsgInstallSC) Route() string {
	return RouterKey
}

func (msg MsgInstallSC) Type() string {
	return "InstallSC"
}

func (msg MsgInstallSC) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	if len(msg.SmartContractID) == 0 {
		return fmt.Errorf("missing smart contract ID")

	}

	if msg.Bytecode == nil || len(msg.Bytecode) == 0 {
		return fmt.Errorf("missing smart contract bytecode")

	}

	return nil
}

func (msg MsgInstallSC) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgInstallSC) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
