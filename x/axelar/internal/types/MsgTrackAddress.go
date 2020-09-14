package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgTrackAddress{}

type MsgTrackAddress struct {
	Sender  sdk.AccAddress
	Chain   string
	Address []byte
}

func NewMsgTrackAddress(sender sdk.AccAddress, chain string, address []byte) MsgTrackAddress {
	return MsgTrackAddress{
		Sender:  sender,
		Chain:   chain,
		Address: address,
	}
}

func (msg MsgTrackAddress) Route() string {
	return RouterKey
}

const TrackAddress = "TrackAddress"

func (msg MsgTrackAddress) Type() string {
	return TrackAddress
}

func (msg MsgTrackAddress) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender must be set")
	}
	if msg.Chain == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "name of the chain for address must be set")
	}
	if msg.Address == nil || len(msg.Address) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address must be set")
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
