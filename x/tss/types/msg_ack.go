package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewHeartBeatRequest constructor for AckRequest
func NewHeartBeatRequest(sender sdk.AccAddress, keyIDs []exported.KeyID) *HeartBeatRequest {
	return &HeartBeatRequest{Sender: sender, KeyIDs: keyIDs}
}

// Route implements the sdk.Msg interface.
func (m HeartBeatRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m HeartBeatRequest) Type() string { return "HeartBeat" }

// ValidateBasic implements the sdk.Msg interface.
func (m HeartBeatRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	for _, keyID := range m.KeyIDs {
		if err := keyID.Validate(); err != nil {
			return sdkerrors.Wrapf(ErrTss, "invalid key ID: %s", err.Error())
		}
	}

	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m HeartBeatRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements the sdk.Msg interface
func (m HeartBeatRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
