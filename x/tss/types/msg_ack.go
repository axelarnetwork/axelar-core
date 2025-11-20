package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHeartBeatRequest constructor for AckRequest
func NewHeartBeatRequest(sender sdk.AccAddress) *HeartBeatRequest {
	// TODO: completely remove keyIDs from the message
	return &HeartBeatRequest{Sender: sender.String(), KeyIDs: nil}
}

// Route implements the sdk.Msg interface.
func (m HeartBeatRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m HeartBeatRequest) Type() string { return "HeartBeat" }

// ValidateBasic implements the sdk.Msg interface.
func (m HeartBeatRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	for _, keyID := range m.KeyIDs {
		if err := keyID.Validate(); err != nil {
			return errorsmod.Wrapf(ErrTss, "invalid key ID: %s", err.Error())
		}
	}

	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m HeartBeatRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
