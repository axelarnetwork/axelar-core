package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewSignPendingTransfersRequest - SignPendingTransfersRequest constructor
func NewSignPendingTransfersRequest(sender sdk.AccAddress, fee btcutil.Amount) *SignPendingTransfersRequest {
	return &SignPendingTransfersRequest{Sender: sender, Fee: int64(fee)}
}

// Route returns the route for this message
func (m SignPendingTransfersRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m SignPendingTransfersRequest) Type() string {
	return "SignPendingTransfers"
}

// ValidateBasic executes a stateless message validation
func (m SignPendingTransfersRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Fee < 0 {
		return fmt.Errorf("fee must be greater than or equal to 0")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m SignPendingTransfersRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m SignPendingTransfersRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
