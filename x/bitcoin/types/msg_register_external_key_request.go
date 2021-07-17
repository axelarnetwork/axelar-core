package types

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterExternalKeyRequest is the constructor for RegisterExternalKeyRequest
func NewRegisterExternalKeyRequest(sender sdk.AccAddress, keyID string, pubKey *btcec.PublicKey) *RegisterExternalKeyRequest {
	return &RegisterExternalKeyRequest{
		Sender: sender,
		KeyID:  keyID,
		PubKey: pubKey.SerializeCompressed(),
	}
}

// Route returns the route for this message
func (m RegisterExternalKeyRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterExternalKeyRequest) Type() string {
	return "RegisterExternalKey"
}

// ValidateBasic executes a stateless message validation
func (m RegisterExternalKeyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.KeyID == "" {
		return sdkerrors.Wrap(ErrBitcoin, "key id must be set")
	}

	if _, err := btcec.ParsePubKey(m.PubKey, btcec.S256()); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterExternalKeyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RegisterExternalKeyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
