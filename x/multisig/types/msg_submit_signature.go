package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &SubmitSignatureRequest{}

// NewSubmitSignatureRequest constructor for SubmitSignatureRequest
func NewSubmitSignatureRequest(sender sdk.AccAddress, sigID uint64, signature Signature) *SubmitSignatureRequest {
	return &SubmitSignatureRequest{
		Sender:    sender,
		SigID:     sigID,
		Signature: signature,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m SubmitSignatureRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Signature.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}

// GetSigners implements the sdk.Msg interface
func (m SubmitSignatureRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
