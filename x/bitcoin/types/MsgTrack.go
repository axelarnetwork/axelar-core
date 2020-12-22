package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgTrack struct {
	Sender  sdk.AccAddress
	Mode    int
	Address BtcAddress
	KeyID   string
	Chain   Chain
	Rescan  bool
}

func NewMsgTrackAddress(sender sdk.AccAddress, address BtcAddress, rescan bool) sdk.Msg {
	return MsgTrack{
		Sender:  sender,
		Address: address,
		Rescan:  rescan,
		Mode:    ModeSpecificAddress,
	}
}

func NewMsgTrackPubKey(sender sdk.AccAddress, chain Chain, keyId string, rescan bool) sdk.Msg {
	return MsgTrack{
		Sender: sender,
		KeyID:  keyId,
		Chain:  chain,
		Rescan: rescan,
		Mode:   ModeSpecificKey,
	}
}

func NewMsgTrackPubKeyWithMasterKey(sender sdk.AccAddress, chain Chain, rescan bool) sdk.Msg {
	return MsgTrack{
		Sender: sender,
		Chain:  chain,
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
	case ModeSpecificKey:
		if msg.KeyID == "" {
			return fmt.Errorf("missing public key ID")
		}
	case ModeCurrentMasterKey:
		if err := msg.Chain.Validate(); err != nil {
			return err
		}
	case ModeSpecificAddress:
		if err := msg.Address.Validate(); err != nil {
			return sdkerrors.Wrap(err, "invalid address to track")
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
