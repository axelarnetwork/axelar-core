package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route returns the route for this message
func (m VotePubKeyRequest) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (m VotePubKeyRequest) Type() string {
	return "VotePubKey"
}

// ValidateBasic performs a stateless validation of this message
func (m VotePubKeyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.PubKeyBytes == nil {
		return fmt.Errorf("missing public key data")
	}
	if _, err := btcec.ParsePubKey(m.PubKeyBytes, btcec.S256()); err != nil {
		return err
	}
	return m.PollMeta.Validate()
}

// GetSignBytes returns the bytes to sign for this message
func (m VotePubKeyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m VotePubKeyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
