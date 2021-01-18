package types

import (
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
)

type MsgLink struct {
	Sender  sdk.AccAddress
	Address exported.CrossChainAddress
}

func NewMsgLink(sender sdk.AccAddress, btcAddress btcutil.Address, destination exported.CrossChainAddress) sdk.Msg {
	return MsgLink{
		Sender:  sender,
		Address: destination,
	}
}
func (msg MsgLink) Route() string {
	return RouterKey
}

func (msg MsgLink) Type() string {
	return "Transfer"
}

func (msg MsgLink) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	if err := msg.Address.Validate(); err != nil {
		return err
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
