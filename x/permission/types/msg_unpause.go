package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewUnpauseRequest is the constructor for UnpauseRequest
func NewUnpauseRequest(sender sdk.AccAddress, reason string) *UnpauseRequest {
	return &UnpauseRequest{
		Sender: sender.String(),
		Reason: reason,
	}
}

// Route returns the route for this message
func (m UnpauseRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m UnpauseRequest) Type() string {
	return "Unpause"
}

// ValidateBasic executes a stateless message validation
func (m UnpauseRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	return validateReason(m.Reason)
}

// GetSignBytes returns the message bytes that need to be signed
func (m UnpauseRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
