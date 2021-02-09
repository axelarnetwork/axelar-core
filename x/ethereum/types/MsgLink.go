package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
)

type MsgLink struct {
	Sender       sdk.AccAddress
	Recipient    exported.CrossChainAddress
	BurneableBC  []byte
	TokenAddr    string
	ContractAddr string
	Salt         [32]byte
}

func NewMsgLink(sender sdk.AccAddress, destination exported.CrossChainAddress, bytecodes []byte, token, contract string, salt [32]byte) sdk.Msg {
	return MsgLink{
		Sender:       sender,
		Recipient:    destination,
		BurneableBC:  bytecodes,
		TokenAddr:    token,
		ContractAddr: contract,
		Salt:         salt,
	}
}

func (msg MsgLink) Route() string {
	return RouterKey
}

func (msg MsgLink) Type() string {
	return "SignPendingTransfersTx"
}

func (msg MsgLink) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

func (msg MsgLink) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgLink) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
