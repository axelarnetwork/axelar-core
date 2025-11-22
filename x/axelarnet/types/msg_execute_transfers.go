package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewExecutePendingTransfersRequest creates a message of type ExecutePendingTransfersRequest
func NewExecutePendingTransfersRequest(sender sdk.AccAddress) *ExecutePendingTransfersRequest {
	return &ExecutePendingTransfersRequest{Sender: sender.String()}
}

// Route returns the route for this message
func (m ExecutePendingTransfersRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m ExecutePendingTransfersRequest) Type() string {
	return "ExecutePendingTransfers"
}

// ValidateBasic executes a stateless message validation
func (m ExecutePendingTransfersRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}
	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m ExecutePendingTransfersRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
