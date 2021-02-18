package types

import (
	"fmt"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// golang stupidity: ensure interface compliance at compile time
var (
	_ sdk.Msg                       = &MsgKeygenStart{}
	_ sdk.Msg                       = MsgAssignNextMasterKey{}
	_ broadcast.MsgWithSenderSetter = &MsgKeygenTraffic{}
	_ broadcast.MsgWithSenderSetter = &MsgSignTraffic{}
	_ voting.MsgVote                = &MsgVotePubKey{}
)

// MsgKeygenStart indicate the start of keygen
type MsgKeygenStart struct {
	Sender    sdk.AccAddress
	NewKeyID  string
	Threshold int
}

// MsgKeygenTraffic protocol message
type MsgKeygenTraffic struct {
	Sender    sdk.AccAddress
	SessionID string
	Payload   *tssd.TrafficOut // pointer because it contains a mutex
}

// MsgSignTraffic protocol message
type MsgSignTraffic struct {
	Sender    sdk.AccAddress
	SessionID string
	Payload   *tssd.TrafficOut // pointer because it contains a mutex
}

// Route implements the sdk.Msg interface.
func (msg MsgKeygenStart) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (msg MsgKeygenStart) Type() string { return "KeyGenStart" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgKeygenStart) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}
	if msg.NewKeyID == "" {
		return sdkerrors.Wrap(ErrTss, "new key id must be set")
	}
	if msg.Threshold < 1 {
		return sdkerrors.Wrap(ErrTss, fmt.Sprintf("invalid threshold [%d]", msg.Threshold))
	}
	// TODO enforce a maximum length for msg.NewKeyID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgKeygenStart) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements the sdk.Msg interface.
func (msg MsgKeygenStart) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
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

// MsgAssignNextMasterKey represents a message to assign a new master key
type MsgAssignNextMasterKey struct {
	Sender sdk.AccAddress
	Chain  string
	KeyID  string
}

// Route returns the route for this message
func (msg MsgAssignNextMasterKey) Route() string { return RouterKey }

// Type returns the type of this message
func (msg MsgAssignNextMasterKey) Type() string { return "AssignNextMasterKey" }

// ValidateBasic performs a stateless validation of this message
func (msg MsgAssignNextMasterKey) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.ErrInvalidAddress
	}
	if msg.KeyID == "" {
		return fmt.Errorf("missing key ID")
	}
	if msg.Chain == "" {
		return fmt.Errorf("missing chain")
	}
	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (msg MsgAssignNextMasterKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgAssignNextMasterKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
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
	if _, err := convert.BytesToPubkey(msg.PubKeyBytes); err != nil {
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

// SetSender sets the sender of this message
func (msg *MsgVotePubKey) SetSender(address sdk.AccAddress) {
	msg.Sender = address
}

// Poll returns the poll this message votes on
func (msg MsgVotePubKey) Poll() voting.PollMeta {
	return msg.PollMeta
}

// Data represents the data value this message votes for
func (msg MsgVotePubKey) Data() voting.VotingData {
	return msg.PubKeyBytes
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
	if _, _, err := convert.BytesToSig(msg.SigBytes); err != nil {
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

// SetSender sets the sender of this message
func (msg *MsgVoteSig) SetSender(address sdk.AccAddress) {
	msg.Sender = address
}

// Poll returns the poll this message votes on
func (msg MsgVoteSig) Poll() voting.PollMeta {
	return msg.PollMeta
}

// Data returns the data value this message votes for
func (msg MsgVoteSig) Data() voting.VotingData {
	return msg.SigBytes
}

// MsgRotateMasterKey represents a message to rotate a master key
type MsgRotateMasterKey struct {
	Sender sdk.AccAddress
	Chain  string
}

// Route returns the route for this message
func (msg MsgRotateMasterKey) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (msg MsgRotateMasterKey) Type() string {
	return "RotateMasterKey"
}

// ValidateBasic performs a stateless validation of this message
func (msg MsgRotateMasterKey) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (msg MsgRotateMasterKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgRotateMasterKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
