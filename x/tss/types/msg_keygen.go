package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewStartKeygenRequest constructor for StartKeygenRequest
func NewStartKeygenRequest(sender sdk.AccAddress, keyID string, keyRole exported.KeyRole, keyType exported.KeyType) *StartKeygenRequest {
	return &StartKeygenRequest{
		Sender: sender.String(),
		KeyInfo: KeyInfo{
			KeyID:   exported.KeyID(keyID),
			KeyRole: keyRole,
			KeyType: keyType,
		},
	}
}

// Route implements the sdk.Msg interface.
func (m StartKeygenRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m StartKeygenRequest) Type() string { return "KeyGenStart" }

// ValidateBasic implements the sdk.Msg interface.
func (m StartKeygenRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.KeyInfo.KeyID.Validate(); err != nil {
		return err
	}

	if err := m.KeyInfo.KeyRole.Validate(); err != nil {
		return err
	}

	if err := m.KeyInfo.KeyType.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements the sdk.Msg interface.
func (m StartKeygenRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// Route implements the sdk.Msg interface.
func (m ProcessKeygenTrafficRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m ProcessKeygenTrafficRequest) Type() string { return "KeygenTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (m ProcessKeygenTrafficRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}
	if m.SessionID == "" {
		return errorsmod.Wrap(ErrTss, "session id must be set")
	}
	if !m.Payload.IsBroadcast && len(m.Payload.ToPartyUid) == 0 {
		return errorsmod.Wrap(ErrTss, "non-broadcast message must specify recipient")
	}
	if m.Payload.IsBroadcast && len(m.Payload.ToPartyUid) != 0 {
		return errorsmod.Wrap(ErrTss, "broadcast message must not specify recipient")
	}
	// TODO enforce a maximum length for m.SessionID?
	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m ProcessKeygenTrafficRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
