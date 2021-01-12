package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgTrack struct {
	Sender  sdk.AccAddress
	Mode    Mode
	Address btcutil.Address
	KeyID   string
	Rescan  bool
}

func NewMsgTrackAddress(sender sdk.AccAddress, address btcutil.Address, rescan bool) sdk.Msg {
	return MsgTrack{
		Sender:  sender,
		Address: address,
		Rescan:  rescan,
		Mode:    ModeSpecificAddress,
	}
}

func NewMsgTrackPubKey(sender sdk.AccAddress, keyId string, rescan bool) sdk.Msg {
	return MsgTrack{
		Sender: sender,
		KeyID:  keyId,
		Rescan: rescan,
		Mode:   ModeSpecificKey,
	}
}

func NewMsgTrackPubKeyWithMasterKey(sender sdk.AccAddress, rescan bool) sdk.Msg {
	return MsgTrack{
		Sender: sender,
		Rescan: rescan,
		Mode:   ModeCurrentMasterKey,
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
	switch msg.Mode {
	case ModeCurrentMasterKey:
		// nothing
	case ModeSpecificKey:
		if msg.KeyID == "" {
			return fmt.Errorf("missing public key ID")
		}
	case ModeSpecificAddress:
		if msg.Address.String() == "" {
			return fmt.Errorf("invalid address to track")
		}
	default:
		return fmt.Errorf("chosen mode not recognized")
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
