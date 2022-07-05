package types

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
)

var _ sdk.Msg = &SubmitPubKeyRequest{}

// NewSubmitPubKeyRequest constructor for AckRequest
func NewSubmitPubKeyRequest(sender sdk.AccAddress, keyID exported.KeyID, pubKey PublicKey, signature []byte) *SubmitPubKeyRequest {
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

	sig, err := btcec.ParseDERSignature(m.Signature, btcec.S256())
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}

// GetSigners implements the sdk.Msg interface
func (m SubmitPubKeyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
