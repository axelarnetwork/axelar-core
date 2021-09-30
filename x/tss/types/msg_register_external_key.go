package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewRegisterExternalKeysRequest is the constructor for RegisterExternalKeysRequest
func NewRegisterExternalKeysRequest(sender sdk.AccAddress, chain string, externalKeys ...RegisterExternalKeysRequest_ExternalKey) *RegisterExternalKeysRequest {
	return &RegisterExternalKeysRequest{
		Sender:       sender,
		Chain:        chain,
		ExternalKeys: externalKeys,
	}
}

// Route returns the route for this message
func (m RegisterExternalKeysRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterExternalKeysRequest) Type() string {
	return "RegisterExternalKey"
}

// ValidateBasic executes a stateless message validation
func (m RegisterExternalKeysRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return sdkerrors.Wrap(ErrTss, "missing chain")
	}

	if len(m.ExternalKeys) == 0 {
		return sdkerrors.Wrap(ErrTss, "no external key is given")
	}

	idMap := make(map[tss.KeyID]bool)
	pubKeyMap := make(map[string]bool)

	for _, externalKey := range m.ExternalKeys {
		if err := externalKey.ID.Validate(); err != nil {
			return err
		}

		if _, err := btcec.ParsePubKey(externalKey.PubKey, btcec.S256()); err != nil {
			return sdkerrors.Wrap(ErrTss, err.Error())
		}

		if idMap[externalKey.ID] {
			return sdkerrors.Wrapf(ErrTss, "duplicate external key id %s found", externalKey.ID)
		}

		pubKeyHex := hex.EncodeToString(externalKey.PubKey)
		if pubKeyMap[pubKeyHex] {
			return sdkerrors.Wrapf(ErrTss, "duplicate external public key %s found", pubKeyHex)
		}

		idMap[externalKey.ID] = true
		pubKeyMap[pubKeyHex] = true
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterExternalKeysRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RegisterExternalKeysRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
