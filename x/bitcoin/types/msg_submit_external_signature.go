package types

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewSubmitExternalSignatureRequest is the constructor for SubmitExternalSignatureRequest
func NewSubmitExternalSignatureRequest(sender sdk.AccAddress, keyID string, signature []byte, sigHash []byte) *SubmitExternalSignatureRequest {
	return &SubmitExternalSignatureRequest{
		Sender:    sender,
		KeyID:     keyID,
		Signature: signature,
		SigHash:   sigHash,
	}
}

// Route returns the route for this message
func (m SubmitExternalSignatureRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m SubmitExternalSignatureRequest) Type() string {
	return "SubmitExternalSignature"
}

// ValidateBasic executes a stateless message validation
func (m SubmitExternalSignatureRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.KeyID == "" {
		return sdkerrors.Wrap(ErrBitcoin, "key id must be set")
	}

	_, err := btcec.ParseDERSignature(m.Signature, btcec.S256())
	if err != nil {
		return sdkerrors.Wrap(ErrBitcoin, err.Error())
	}

	if len(m.SigHash) == 0 {
		return sdkerrors.Wrap(ErrBitcoin, "sig hash must be set")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m SubmitExternalSignatureRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m SubmitExternalSignatureRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
