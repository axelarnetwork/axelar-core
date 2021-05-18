package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeregisterRequest creates a message of type DeregisterRequest
func NewDeregisterRequest(sender sdk.AccAddress) *DeregisterRequest {
	return &DeregisterRequest{
		Sender: sender,
	}
}

// Route implements sdk.Msg
func (m DeregisterRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m DeregisterRequest) Type() string {
	return "Deregister"
}

// ValidateBasic implements sdk.Msg
func (m DeregisterRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m DeregisterRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m DeregisterRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
