package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewCreatePendingTransfersTxRequest - CreatePendingTransfersTxRequest constructor
func NewCreatePendingTransfersTxRequest(sender sdk.AccAddress, keyID string, masterKeyAmount btcutil.Amount) *CreatePendingTransfersTxRequest {
	return &CreatePendingTransfersTxRequest{
		Sender:          sender,
		KeyID:           tss.KeyID(keyID),
		MasterKeyAmount: masterKeyAmount,
	}
}

// Route returns the route for this message
func (m CreatePendingTransfersTxRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m CreatePendingTransfersTxRequest) Type() string {
	return "CreatePendingTransfersTx"
}

// ValidateBasic executes a stateless message validation
func (m CreatePendingTransfersTxRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.KeyID.Validate(); err != nil {
		return err
	}

	if m.MasterKeyAmount < 0 {
		return fmt.Errorf("master key amount must be >= 0")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m CreatePendingTransfersTxRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m CreatePendingTransfersTxRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
