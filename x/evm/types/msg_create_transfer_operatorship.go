package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCreateTransferOperatorshipRequest creates a message of type CreateTransferOperatorshipRequest
func NewCreateTransferOperatorshipRequest(sender sdk.AccAddress, chain string, keyID string) *CreateTransferOperatorshipRequest {
	return &CreateTransferOperatorshipRequest{
		Sender: sender,
		Chain:  chain,
		KeyID:  keyID,
	}
}

// Route implements sdk.Msg
func (m CreateTransferOperatorshipRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m CreateTransferOperatorshipRequest) Type() string {
	return "CreateTransferOperatorship"
}

// ValidateBasic implements sdk.Msg
func (m CreateTransferOperatorshipRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if m.KeyID == "" {
		return fmt.Errorf("missing key ID")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m CreateTransferOperatorshipRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m CreateTransferOperatorshipRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
