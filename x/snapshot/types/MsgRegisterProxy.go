package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterProxyRequest - RegisterProxyRequest constructor
func NewRegisterProxyRequest(sender sdk.AccAddress, proxy sdk.AccAddress) *RegisterProxyRequest {
	return &RegisterProxyRequest{
		Sender:    sender.String(),
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "principal").Error())
	}
	if err := sdk.VerifyAddressFormat(m.ProxyAddr); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "proxy").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterProxyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
