package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &SubmitSignatureRequest{}

// NewSubmitSignatureRequest constructor for SubmitSignatureRequest
func NewSubmitSignatureRequest(sender sdk.AccAddress, sigID uint64, signature Signature) *SubmitSignatureRequest {
	return &SubmitSignatureRequest{
		Sender:    sender.String(),
		SigID:     sigID,
		Signature: signature,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m SubmitSignatureRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Signature.ValidateBasic(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}
