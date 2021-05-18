package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route implements the sdk.Msg interface.
func (m ProcessSignTrafficRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m ProcessSignTrafficRequest) Type() string { return "SignTraffic" }

// ValidateBasic implements the sdk.Msg interface.
func (m ProcessSignTrafficRequest) ValidateBasic() error {
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
func (m ProcessSignTrafficRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements the sdk.Msg interface
func (m ProcessSignTrafficRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
