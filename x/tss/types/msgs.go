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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed sender address")
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
	return []sdk.AccAddress{m.GetSender()}
}

// SetSender implements the broadcast.MsgWithSenderSetter interface
func (m *MsgKeygenTraffic) SetSender(sender sdk.AccAddress) {
	m.Sender = sender.String()
}

// GetSender returns the sender object
func (m MsgKeygenTraffic) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return addr
}

// Route implements the sdk.Msg interface.
func (m MsgSignTraffic) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m MsgSignTraffic) Type() string { return "SignTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (m MsgSignTraffic) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed sender address")
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
	return []sdk.AccAddress{m.GetSender()}
}

// SetSender implements the broadcast.MsgWithSenderSetter interface
func (m *MsgSignTraffic) SetSender(sender sdk.AccAddress) {
	m.Sender = sender.String()
}

// GetSender returns the sender object
func (m MsgSignTraffic) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return addr
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed sender address")
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
	return []sdk.AccAddress{m.GetSender()}
}

// GetSender returns the sender object
func (m MsgVotePubKey) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return addr
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed sender address")
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
	return []sdk.AccAddress{m.GetSender()}
}

// GetSender returns the sender object
func (m MsgVoteSig) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return addr
}
