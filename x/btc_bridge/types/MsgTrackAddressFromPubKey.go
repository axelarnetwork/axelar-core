package types

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgTrackAddressFromPubKey{}

type MsgTrackAddressFromPubKey struct {
	Sender sdk.AccAddress
	KeyID  string
	Chain  string
}

func NewMsgTrackAddressFromPubKey(sender sdk.AccAddress, keyId string, chain string) MsgTrackAddressFromPubKey {
	return MsgTrackAddressFromPubKey{
		Sender: sender,
		KeyID:  keyId,
		Chain:  chain,
	}
}

func (msg MsgTrackAddressFromPubKey) Route() string {
	return RouterKey
}

func (msg MsgTrackAddressFromPubKey) Type() string {
	return "TrackAddress"
}

func (msg MsgTrackAddressFromPubKey) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.KeyID == "" {
		return fmt.Errorf("missing public key ID")
	}
	switch msg.Chain {
	case chaincfg.MainNetParams.Name, chaincfg.TestNet3Params.Name:
		break
	default:
		return fmt.Errorf(
			"missing chain name, choose %s or %s",
			chaincfg.MainNetParams.Name,
			chaincfg.TestNet3Params.Name,
		)
	}

	return nil
}

func (msg MsgTrackAddressFromPubKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgTrackAddressFromPubKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
