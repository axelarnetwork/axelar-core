package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterControllerRequest is the constructor for RegisterControllerRequest
func NewRegisterControllerRequest(sender sdk.AccAddress, controller sdk.AccAddress) *RegisterControllerRequest {
	return &RegisterControllerRequest{
		Sender:     sender.String(),
		Controller: controller,
	}
}

// Route returns the route for this message
func (m RegisterControllerRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterControllerRequest) Type() string {
	return "RegisterController"
}

// ValidateBasic executes a stateless message validation
func (m RegisterControllerRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := sdk.VerifyAddressFormat(m.Controller); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "controller").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterControllerRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
