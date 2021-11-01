package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRouteIBCTransfersRequest creates a message of type RouteIBCTransfersRequest
func NewRouteIBCTransfersRequest(sender sdk.AccAddress) *RouteIBCTransfersRequest {
	return &RouteIBCTransfersRequest{Sender: sender}
}

// Route returns the route for this message
func (m RouteIBCTransfersRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RouteIBCTransfersRequest) Type() string {
	return "RouteIBCTransfers"
}

// ValidateBasic executes a stateless message validation
func (m RouteIBCTransfersRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RouteIBCTransfersRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m RouteIBCTransfersRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
