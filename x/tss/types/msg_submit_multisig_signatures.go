package types

import (
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewSubmitMultisigSignaturesRequest constructor for SubmitMultisigSignaturesRequest
func NewSubmitMultisigSignaturesRequest(sender sdk.AccAddress, sigID string, signatures [][]byte) *SubmitMultisigSignaturesRequest {
	return &SubmitMultisigSignaturesRequest{Sender: sender, SigID: utils.NormalizeString(sigID), Signatures: signatures}
}

// Route implements the sdk.Msg interface.
func (m SubmitMultisigSignaturesRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m SubmitMultisigSignaturesRequest) Type() string { return "SubmitMultisigSignatures" }

// ValidateBasic implements the sdk.Msg interface.
func (m SubmitMultisigSignaturesRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.SigID); err != nil {
		return sdkerrors.Wrap(err, "invalid signature ID")
	}

	for _, sig := range m.Signatures {
		_, err := btcec.ParseDERSignature(sig, btcec.S256())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSignBytes implements the sdk.Msg interface
func (m SubmitMultisigSignaturesRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements the sdk.Msg interface
func (m SubmitMultisigSignaturesRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
