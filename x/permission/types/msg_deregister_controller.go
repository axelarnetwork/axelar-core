package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeregisterControllerRequest is the constructor for DeregisterControllerRequest
func NewDeregisterControllerRequest(sender sdk.AccAddress, controller sdk.AccAddress) *DeregisterControllerRequest {
	return &DeregisterControllerRequest{
		Sender:     sender.String(),
		Controller: controller,
	}
}

// Route returns the route for this message
func (m DeregisterControllerRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m DeregisterControllerRequest) Type() string {
	return "DeregisterController"
}

// ValidateBasic executes a stateless message validation
func (m DeregisterControllerRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := sdk.VerifyAddressFormat(m.Controller); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "controller").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m DeregisterControllerRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
