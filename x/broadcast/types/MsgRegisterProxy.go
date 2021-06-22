package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeregisterProxyRequest - DeregisterProxyRequest constructor
func NewDeregisterProxyRequest(principal sdk.ValAddress) *DeregisterProxyRequest {
	return &DeregisterProxyRequest{
		PrincipalAddr: principal,
	}
}

// Route returns the route for this message
func (m DeregisterProxyRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m DeregisterProxyRequest) Type() string {
	return "DeregisterProxy"
}

// ValidateBasic executes a stateless message validation
func (m DeregisterProxyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.PrincipalAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "principal").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m DeregisterProxyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m DeregisterProxyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(m.PrincipalAddr)}
}
