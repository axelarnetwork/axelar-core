package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRetryIBCTransferRequest creates a message of type RetryIBCTransferRequest
func NewRetryIBCTransferRequest(sender sdk.AccAddress, id nexus.TransferID) *RetryIBCTransferRequest {
	return &RetryIBCTransferRequest{
		Sender: sender,
		Chain:  nexus.ChainName(""),
		ID:     id,
	}
}

// Route returns the route for this message
func (m RetryIBCTransferRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RetryIBCTransferRequest) Type() string {
	return "RetryIBCTransfer"
}

// ValidateBasic executes a stateless message validation
func (m RetryIBCTransferRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RetryIBCTransferRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m RetryIBCTransferRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
