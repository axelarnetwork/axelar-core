package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCreatePendingTransfersRequest - CreatePendingTransfersRequest constructor
func NewCreatePendingTransfersRequest(sender sdk.AccAddress, chain string) *CreatePendingTransfersRequest {
	return &CreatePendingTransfersRequest{Sender: sender, Chain: chain}
}

// Route returns the route for this message
func (m CreatePendingTransfersRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m CreatePendingTransfersRequest) Type() string {
	return "CreatePendingTransfers"
}

// ValidateBasic executes a stateless message validation
func (m CreatePendingTransfersRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m CreatePendingTransfersRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m CreatePendingTransfersRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
