package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
)

type MsgTransfer struct {
	Sender      sdk.AccAddress
	BTCAddress  btcutil.Address
	Destination exported.CrossChainAddress
	Amount      btcutil.Amount
}

func NewMsgTransfer(sender sdk.AccAddress, btcAddress btcutil.Address, destination exported.CrossChainAddress) sdk.Msg {
	return MsgTransfer{
		Sender:      sender,
		BTCAddress:  btcAddress,
		Destination: destination,
	}
}
func (msg MsgTransfer) Route() string {
	return RouterKey
}

func (msg MsgTransfer) Type() string {
	return "Transfer"
}

func (msg MsgTransfer) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if err := msg.Destination.Validate(); err != nil {
		return err
	}
	if msg.BTCAddress.String() == "" {
		return fmt.Errorf("invalid address to track")
	}
	return nil
}

func (msg MsgTransfer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
