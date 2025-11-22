package types

import (
	errorsmod "cosmossdk.io/errors"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// NewSubmitMultisigSignaturesRequest constructor for SubmitMultisigSignaturesRequest
func NewSubmitMultisigSignaturesRequest(sender sdk.AccAddress, sigID string, signatures [][]byte) *SubmitMultisigSignaturesRequest {
	return &SubmitMultisigSignaturesRequest{Sender: sender.String(), SigID: utils.NormalizeString(sigID), Signatures: signatures}
}

// Route implements the sdk.Msg interface.
func (m SubmitMultisigSignaturesRequest) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
// naming convention follows x/staking/types/msgs.go
func (m SubmitMultisigSignaturesRequest) Type() string { return "SubmitMultisigSignatures" }

// ValidateBasic implements the sdk.Msg interface.
func (m SubmitMultisigSignaturesRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.SigID); err != nil {
		return errorsmod.Wrap(err, "invalid signature ID")
	}

	if len(m.Signatures) == 0 {
		return errorsmod.Wrap(ErrTss, "no signature is given")
	}

	for _, sig := range m.Signatures {
		_, err := ec.ParseDERSignature(sig)
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
