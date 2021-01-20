package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgTrack struct {
	Sender  sdk.AccAddress
	Address btcutil.Address
	Rescan  bool
}

func NewMsgTrackAddress(sender sdk.AccAddress, address btcutil.Address, rescan bool) sdk.Msg {
	return MsgTrack{
		Sender:  sender,
		Address: address,
		Rescan:  rescan,
	}
}

func (msg MsgTrack) Route() string {
	return RouterKey
}

func (msg MsgTrack) Type() string {
	return "Track"
}

func (msg MsgTrack) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Address.String() == "" {
		return fmt.Errorf("invalid address to track")
	}
	return nil
}

func (msg MsgTrack) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTrack) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
