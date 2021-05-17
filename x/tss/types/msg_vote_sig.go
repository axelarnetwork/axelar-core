package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route returns the route for this message
func (m VoteSigRequest) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (m VoteSigRequest) Type() string {
	return "VoteSig"
}

// ValidateBasic performs a stateless validation of this message
func (m VoteSigRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.SigBytes == nil {
		return fmt.Errorf("missing signature data")
	}
	if _, err := btcec.ParseDERSignature(m.SigBytes, btcec.S256()); err != nil {
		return err
	}
	return m.PollMeta.Validate()
}

// GetSignBytes returns the bytes to sign for this message
func (m VoteSigRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m VoteSigRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
