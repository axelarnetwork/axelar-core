package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewSignMasterConsolidationTransactionRequest is the constructor for SignMasterConsolidationTransactionRequest
func NewSignMasterConsolidationTransactionRequest(sender sdk.AccAddress, keyID string, secondaryKeyAmount btcutil.Amount) *SignMasterConsolidationTransactionRequest {
	return &SignMasterConsolidationTransactionRequest{
		Sender:             sender,
		KeyID:              keyID,
		SecondaryKeyAmount: secondaryKeyAmount,
	}
}

// Route returns the route for this message
func (m SignMasterConsolidationTransactionRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m SignMasterConsolidationTransactionRequest) Type() string {
	return "SignMasterConsolidationTransaction"
}

// ValidateBasic executes a stateless message validation
func (m SignMasterConsolidationTransactionRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.KeyID == "" {
		return sdkerrors.Wrap(ErrBitcoin, "key id must be set")
	}

	if m.SecondaryKeyAmount < 0 {
		return fmt.Errorf("secondary key amount must be >= 0")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m SignMasterConsolidationTransactionRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m SignMasterConsolidationTransactionRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
