package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCreateRescueTxRequest is the constructor for CreateRescueTxRequest
func NewCreateRescueTxRequest(sender sdk.AccAddress) *CreateRescueTxRequest {
	return &CreateRescueTxRequest{
		Sender: sender,
	}
}

// Route returns the route for this message
func (m CreateRescueTxRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m CreateRescueTxRequest) Type() string {
	return "SignRescueTransaction"
}

// ValidateBasic executes a stateless message validation
func (m CreateRescueTxRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m CreateRescueTxRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m CreateRescueTxRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
