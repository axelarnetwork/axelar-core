package types

import (
	"crypto/sha256"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
)

var _ sdk.Msg = &SubmitPubKeyRequest{}

// NewSubmitPubKeyRequest constructor for SubmitPubKeyRequest
func NewSubmitPubKeyRequest(sender sdk.AccAddress, keyID exported.KeyID, pubKey exported.PublicKey, signature Signature) *SubmitPubKeyRequest {
	return &SubmitPubKeyRequest{
		Sender:    sender.String(),
		KeyID:     keyID,
		PubKey:    pubKey,
		Signature: signature,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m SubmitPubKeyRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.KeyID.ValidateBasic(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := m.PubKey.ValidateBasic(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := m.Signature.ValidateBasic(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	hash := sha256.Sum256([]byte(m.Sender))
	if !m.Signature.Verify(hash[:], m.PubKey) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "signature does not match the public key")
	}

	return nil
}
