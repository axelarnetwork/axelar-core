package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeactivateProxyRequest - DeregisterProxyRequest constructor
func NewDeactivateProxyRequest(sender sdk.AccAddress) *DeactivateProxyRequest {
	return &DeactivateProxyRequest{
		Sender: sender.String(),
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "principal").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m DeactivateProxyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}
