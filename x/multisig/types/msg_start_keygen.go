package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &StartKeygenRequest{}

// NewStartKeygenRequest constructor for StartKeygenRequest
func NewStartKeygenRequest(sender sdk.AccAddress) *StartKeygenRequest {
	return &StartKeygenRequest{
		Sender: sender,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m StartKeygenRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSigners implements the sdk.Msg interface
func (m StartKeygenRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
