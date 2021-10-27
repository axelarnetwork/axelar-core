package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewProxyReadyRequest - ProxyReadyRequest constructor
func NewProxyReadyRequest(sender sdk.AccAddress, operator sdk.ValAddress) *ProxyReadyRequest {
	return &ProxyReadyRequest{
		Sender:       sender,
		OperatorAddr: operator,
	}
}

// Route returns the route for this message
func (m ProxyReadyRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m ProxyReadyRequest) Type() string {
	return "ProxyReady"
}

// ValidateBasic executes a stateless message validation
func (m ProxyReadyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if err := sdk.VerifyAddressFormat(m.OperatorAddr); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "operator").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m ProxyReadyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m ProxyReadyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(m.Sender)}
}
