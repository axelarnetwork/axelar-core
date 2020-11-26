package types

import (
	"fmt"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	voting "github.com/axelarnetwork/axelar-core/x/voting/exported"
)

// golang stupidity: ensure interface compliance at compile time
var (
	_ sdk.Msg                       = &MsgKeygenStart{}
	_ sdk.Msg                       = &MsgSignStart{}
	_ sdk.Msg                       = MsgMasterKeyRefresh{}
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

// MsgSignStart indicate the start of sign
type MsgSignStart struct {
	Sender    sdk.AccAddress
	NewSigID  string
	KeyID     string
	MsgToSign []byte
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
func (msg MsgKeygenStart) Type() string { return "keygen_start" }

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
	// TODO enforce a maximum length for msg.SessionID?
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
func (msg MsgSignStart) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (msg MsgSignStart) Type() string { return "sign_start" }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgSignStart) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}
	if msg.NewSigID == "" {
		return sdkerrors.Wrap(ErrTss, "new sig id must be set")
	}
	if msg.KeyID == "" {
		return sdkerrors.Wrap(ErrTss, "key id must be set")
	}
	if msg.MsgToSign == nil {
		return sdkerrors.Wrap(ErrTss, "msg must be set")
	}
	// TODO enforce a maximum length for msg.SessionID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgSignStart) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements the sdk.Msg interface.
func (msg MsgSignStart) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// Route implements the sdk.Msg interface.
func (msg MsgKeygenTraffic) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (msg MsgKeygenTraffic) Type() string { return "in" }

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
func (msg MsgSignTraffic) Type() string { return "in" }

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

type MsgMasterKeyRefresh struct {
	Sender sdk.AccAddress
	Chain  string
}

func (msg MsgMasterKeyRefresh) Route() string { return RouterKey }
func (msg MsgMasterKeyRefresh) Type() string  { return "MasterKeyRefresh" }

func (msg MsgMasterKeyRefresh) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.ErrInvalidAddress
	}
	return nil
}

func (msg MsgMasterKeyRefresh) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgMasterKeyRefresh) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

type MsgVotePubKey struct {
	Sender   sdk.AccAddress
	PollMeta voting.PollMeta
	// need to vote on the bytes instead of ecds.PublicKey, otherwise we lose the elliptic curve information
	PubKeyBytes []byte
}

func (msg MsgVotePubKey) Route() string {
	return RouterKey
}

func (msg MsgVotePubKey) Type() string {
	return msg.Type()
}

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
	return nil
}

func (msg MsgVotePubKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVotePubKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg *MsgVotePubKey) SetSender(address sdk.AccAddress) {
	msg.Sender = address
}

func (msg MsgVotePubKey) Poll() voting.PollMeta {
	return msg.PollMeta
}

func (msg MsgVotePubKey) Data() voting.VotingData {
	return msg.PubKeyBytes
}
