package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterProxyRequest - RegisterProxyRequest constructor
func NewRegisterProxyRequest(sender sdk.ValAddress, proxy sdk.AccAddress) *RegisterProxyRequest {
	return &RegisterProxyRequest{
		Sender:    sender,
		ProxyAddr: proxy,
	}
}

// Route returns the route for this message
func (m RegisterProxyRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterProxyRequest) Type() string {
	return "RegisterProxy"
}

// ValidateBasic executes a stateless message validation
func (m RegisterProxyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "principal").Error())
	}
	if err := sdk.VerifyAddressFormat(m.ProxyAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "proxy").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterProxyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RegisterProxyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(m.Sender)}
}
