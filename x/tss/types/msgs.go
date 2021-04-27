package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route implements the sdk.Msg interface.
func (m MsgKeygenTraffic) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m MsgKeygenTraffic) Type() string { return "KeygenTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (m MsgKeygenTraffic) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.SessionID == "" {
		return sdkerrors.Wrap(ErrTss, "session id must be set")
	}
	if !m.Payload.IsBroadcast && len(m.Payload.ToPartyUid) == 0 {
		return sdkerrors.Wrap(ErrTss, "non-broadcast message must specify recipient")
	}
	if m.Payload.IsBroadcast && len(m.Payload.ToPartyUid) != 0 {
		return sdkerrors.Wrap(ErrTss, "broadcast message must not specify recipient")
	}
	// TODO enforce a maximum length for m.SessionID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m MsgKeygenTraffic) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements the sdk.Msg interface
func (m MsgKeygenTraffic) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// Route implements the sdk.Msg interface.
func (m MsgSignTraffic) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m MsgSignTraffic) Type() string { return "SignTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (m MsgSignTraffic) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.SessionID == "" {
		return sdkerrors.Wrap(ErrTss, "session id must be set")
	}
	if !m.Payload.IsBroadcast && len(m.Payload.ToPartyUid) == 0 {
		return sdkerrors.Wrap(ErrTss, "non-broadcast message must specify recipient")
	}
	if m.Payload.IsBroadcast && len(m.Payload.ToPartyUid) != 0 {
		return sdkerrors.Wrap(ErrTss, "broadcast message must not specify recipient")
	}
	// TODO enforce a maximum length for m.SessionID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m MsgSignTraffic) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements the sdk.Msg interface
func (m MsgSignTraffic) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// Route returns the route for this message
func (m MsgVotePubKey) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (m MsgVotePubKey) Type() string {
	return "VotePubKey"
}

// ValidateBasic performs a stateless validation of this message
func (m MsgVotePubKey) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.PubKeyBytes == nil {
		return fmt.Errorf("missing public key data")
	}
	if _, err := btcec.ParsePubKey(m.PubKeyBytes, btcec.S256()); err != nil {
		return err
	}
	return m.PollMeta.Validate()
}

// GetSignBytes returns the bytes to sign for this message
func (m MsgVotePubKey) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgVotePubKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// Route returns the route for this message
func (m MsgVoteSig) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (m MsgVoteSig) Type() string {
	return "VoteSig"
}

// ValidateBasic performs a stateless validation of this message
func (m MsgVoteSig) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.SigBytes == nil {
		return fmt.Errorf("missing signature data")
	}
	if _, err := btcec.ParseDERSignature(m.SigBytes, btcec.S256()); err != nil {
		return err
	}
	return m.PollMeta.Validate()
}

// GetSignBytes returns the bytes to sign for this message
func (m MsgVoteSig) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgVoteSig) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
