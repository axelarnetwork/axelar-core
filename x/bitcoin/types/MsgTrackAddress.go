package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgTrackAddress struct {
	Sender  sdk.AccAddress
	Address BtcAddress
	Rescan  bool
}

func NewMsgTrackAddress(sender sdk.AccAddress, address BtcAddress, rescan bool) sdk.Msg {
	return MsgTrackAddress{
		Sender:  sender,
		Address: address,
		Rescan:  rescan,
	}
}

func (msg MsgTrackAddress) Route() string {
	return RouterKey
}

func (msg MsgTrackAddress) Type() string {
	return "TrackAddress"
}

func (msg MsgTrackAddress) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if err := msg.Address.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid address to track")
	}

	return nil
}

func (msg MsgTrackAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTrackAddress) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
