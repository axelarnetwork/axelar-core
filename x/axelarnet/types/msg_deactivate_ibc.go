package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeactivateIBCRequest creates a message of type DeactivateIBCRequest
func NewDeactivateIBCRequest(sender sdk.AccAddress) *DeactivateIBCRequest {
	return &DeactivateIBCRequest{
		Sender: sender,
	}
}

// Route implements sdk.Msg
func (m DeactivateIBCRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m DeactivateIBCRequest) Type() string {
	return "DeactivateIBC"
}

// ValidateBasic implements sdk.Msg
func (m DeactivateIBCRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m DeactivateIBCRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m DeactivateIBCRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
