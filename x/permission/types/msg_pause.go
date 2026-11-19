package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// NewPauseRequest is the constructor for PauseRequest
func NewPauseRequest(sender sdk.AccAddress, level exported.PauseLevel, reason string) *PauseRequest {
	return &PauseRequest{
		Sender: sender.String(),
		Level:  level,
		Reason: reason,
	}
}

// Route returns the route for this message
func (m PauseRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m PauseRequest) Type() string {
	return "Pause"
}

// ValidateBasic executes a stateless message validation
func (m PauseRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Level.ValidatePauseLevel(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return validateReason(m.Reason)
}

// GetSignBytes returns the message bytes that need to be signed
func (m PauseRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

func validateReason(reason string) error {
	if reason == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "reason must not be empty")
	}

	return nil
}
