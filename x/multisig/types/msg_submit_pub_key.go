package types

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
)

var _ sdk.Msg = &SubmitPubKeyRequest{}

// NewSubmitPubKeyRequest constructor for SubmitPubKeyRequest
func NewSubmitPubKeyRequest(sender sdk.AccAddress, keyID exported.KeyID, pubKey exported.PublicKey, signature Signature) *SubmitPubKeyRequest {
	return &SubmitPubKeyRequest{
		Sender:    sender,
		KeyID:     keyID,
		PubKey:    pubKey,
		Signature: signature,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m SubmitPubKeyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.KeyID.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := m.PubKey.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := m.Signature.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	hash := sha256.Sum256([]byte(m.Sender))
	if !m.Signature.Verify(hash[:], m.PubKey) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "signature does not match the public key")
	}

	return nil
}

// GetSigners implements the sdk.Msg interface
func (m SubmitPubKeyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
