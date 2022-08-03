package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRetryFailedTransferRequest creates a message of type RetryFailedTransferRequest
func NewRetryFailedTransferRequest(sender sdk.AccAddress, id uint64) *RetryFailedTransferRequest {
	return &RetryFailedTransferRequest{
		Sender: sender,
		ID:     nexus.TransferID(id),
	}
}

// Route returns the route for this message
func (m RetryFailedTransferRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RetryFailedTransferRequest) Type() string {
	return "RetryFailedTransfer"
}

// ValidateBasic executes a stateless message validation
func (m RetryFailedTransferRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RetryFailedTransferRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m RetryFailedTransferRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
