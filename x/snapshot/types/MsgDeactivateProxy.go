package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeactivateProxyRequest - DeregisterProxyRequest constructor
func NewDeactivateProxyRequest(principal sdk.ValAddress) *DeactivateProxyRequest {
	return &DeactivateProxyRequest{
		PrincipalAddr: principal,
	}
}

// Route returns the route for this message
func (m DeactivateProxyRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m DeactivateProxyRequest) Type() string {
	return "DeregisterProxy"
}

// ValidateBasic executes a stateless message validation
func (m DeactivateProxyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.PrincipalAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "principal").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m DeactivateProxyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m DeactivateProxyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(m.PrincipalAddr)}
}
