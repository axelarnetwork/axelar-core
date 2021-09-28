package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewCreateMasterTxRequest is the constructor for CreateMasterTxRequest
func NewCreateMasterTxRequest(sender sdk.AccAddress, keyID string, secondaryKeyAmount btcutil.Amount) *CreateMasterTxRequest {
	return &CreateMasterTxRequest{
		Sender:             sender,
		KeyID:              tss.KeyID(keyID),
		SecondaryKeyAmount: secondaryKeyAmount,
	}
}

// Route returns the route for this message
func (m CreateMasterTxRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m CreateMasterTxRequest) Type() string {
	return "SignMasterConsolidationTransaction"
}

// ValidateBasic executes a stateless message validation
func (m CreateMasterTxRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.KeyID.Validate(); err != nil {
		return err
	}

	if m.SecondaryKeyAmount < 0 {
		return fmt.Errorf("secondary key amount must be >= 0")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m CreateMasterTxRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m CreateMasterTxRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
