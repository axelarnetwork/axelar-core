package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type MsgTrackPubKey struct {
	Sender       sdk.AccAddress
	KeyID        string
	UseMasterKey bool
	Chain        Chain
	Rescan       bool
}

func NewMsgTrackPubKeyWithMasterKey(sender sdk.AccAddress, chain Chain, rescan bool) sdk.Msg {
	return MsgTrackPubKey{
		Sender:       sender,
		Chain:        chain,
		Rescan:       rescan,
		UseMasterKey: true,
	}
}

func NewMsgTrackPubKey(sender sdk.AccAddress, chain Chain, keyId string, rescan bool) sdk.Msg {
	return MsgTrackPubKey{
		Sender: sender,
		KeyID:  keyId,
		Chain:  chain,
		Rescan: rescan,
	}
}

func (msg MsgTrackPubKey) Route() string {
	return RouterKey
}

func (msg MsgTrackPubKey) Type() string {
	return "TrackPubKey"
}

func (msg MsgTrackPubKey) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.KeyID == "" && !msg.UseMasterKey {
		return fmt.Errorf("missing public key ID")
	}
	if err := msg.Chain.Validate(); err != nil {
		return err
	}

	return nil
}

func (msg MsgTrackPubKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTrackPubKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
