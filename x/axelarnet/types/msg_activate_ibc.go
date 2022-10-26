package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewActivateIBCRequest creates a message of type ActivateIBCRequest
func NewActivateIBCRequest(sender sdk.AccAddress) *ActivateIBCRequest {
	return &ActivateIBCRequest{
		Sender: sender,
	}
}

// Route implements sdk.Msg
func (m ActivateIBCRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ActivateIBCRequest) Type() string {
	return "ActivateIBC"
}

// ValidateBasic implements sdk.Msg
func (m ActivateIBCRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ActivateIBCRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ActivateIBCRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
