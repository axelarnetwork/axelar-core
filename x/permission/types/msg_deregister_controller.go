package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeregisterControllerRequest is the constructor for DeregisterControllerRequest
func NewDeregisterControllerRequest(sender sdk.AccAddress, controller sdk.AccAddress) *DeregisterControllerRequest {
	return &DeregisterControllerRequest{
		Sender:     sender,
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
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := sdk.VerifyAddressFormat(m.Controller); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "controller").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m DeregisterControllerRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m DeregisterControllerRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
