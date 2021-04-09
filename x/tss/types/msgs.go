package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// MsgKeygenTraffic protocol message
type MsgKeygenTraffic struct {
	Sender    sdk.AccAddress
	SessionID string
	Payload   *tofnd.TrafficOut // pointer because it contains a mutex
}

// MsgSignTraffic protocol message
type MsgSignTraffic struct {
	Sender    sdk.AccAddress
	SessionID string
	Payload   *tofnd.TrafficOut // pointer because it contains a mutex
}

// Route implements the sdk.Msg interface.
func (msg MsgKeygenTraffic) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (msg MsgKeygenTraffic) Type() string { return "KeygenTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgKeygenTraffic) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}
	if msg.SessionID == "" {
		return sdkerrors.Wrap(ErrTss, "session id must be set")
	}
	if !msg.Payload.IsBroadcast && len(msg.Payload.ToPartyUid) == 0 {
		return sdkerrors.Wrap(ErrTss, "non-broadcast message must specify recipient")
	}
	if msg.Payload.IsBroadcast && len(msg.Payload.ToPartyUid) != 0 {
		return sdkerrors.Wrap(ErrTss, "broadcast message must not specify recipient")
	}
	// TODO enforce a maximum length for msg.SessionID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgKeygenTraffic) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements the sdk.Msg interface
func (msg MsgKeygenTraffic) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// SetSender implements the broadcast.MsgWithSenderSetter interface
func (msg *MsgKeygenTraffic) SetSender(sender sdk.AccAddress) {
	msg.Sender = sender
}

// Route implements the sdk.Msg interface.
func (msg MsgSignTraffic) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (msg MsgSignTraffic) Type() string { return "SignTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgSignTraffic) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}
	if msg.SessionID == "" {
		return sdkerrors.Wrap(ErrTss, "session id must be set")
	}
	if !msg.Payload.IsBroadcast && len(msg.Payload.ToPartyUid) == 0 {
		return sdkerrors.Wrap(ErrTss, "non-broadcast message must specify recipient")
	}
	if msg.Payload.IsBroadcast && len(msg.Payload.ToPartyUid) != 0 {
		return sdkerrors.Wrap(ErrTss, "broadcast message must not specify recipient")
	}
	// TODO enforce a maximum length for msg.SessionID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgSignTraffic) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements the sdk.Msg interface
func (msg MsgSignTraffic) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// SetSender implements the broadcast.MsgWithSenderSetter interface
func (msg *MsgSignTraffic) SetSender(sender sdk.AccAddress) {
	msg.Sender = sender
}

// MsgVotePubKey represents the message to vote on a public key
type MsgVotePubKey struct {
	Sender   sdk.AccAddress
	PollMeta voting.PollMeta
	// need to vote on the bytes instead of ecdsa.PublicKey, otherwise we lose the elliptic curve information
	PubKeyBytes []byte
}

// Route returns the route for this message
func (msg MsgVotePubKey) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (msg MsgVotePubKey) Type() string {
	return "VotePubKey"
}

// ValidateBasic performs a stateless validation of this message
func (msg MsgVotePubKey) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.PubKeyBytes == nil {
		return fmt.Errorf("missing public key data")
	}
	if _, err := btcec.ParsePubKey(msg.PubKeyBytes, btcec.S256()); err != nil {
		return err
	}
	return msg.PollMeta.Validate()
}

// GetSignBytes returns the bytes to sign for this message
func (msg MsgVotePubKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgVotePubKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgVoteSig represents a message to vote for a signature
type MsgVoteSig struct {
	Sender   sdk.AccAddress
	PollMeta voting.PollMeta
	// need to vote on the bytes instead of r, s, because Go cannot deserialize private fields using reflection (so *big.Int does not work)
	SigBytes []byte
}

// Route returns the route for this message
func (msg MsgVoteSig) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (msg MsgVoteSig) Type() string {
	return "VoteSig"
}

// ValidateBasic performs a stateless validation of this message
func (msg MsgVoteSig) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.SigBytes == nil {
		return fmt.Errorf("missing signature data")
	}
	if _, err := btcec.ParseDERSignature(msg.SigBytes, btcec.S256()); err != nil {
		return err
	}
	return msg.PollMeta.Validate()
}

// GetSignBytes returns the bytes to sign for this message
func (msg MsgVoteSig) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgVoteSig) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
