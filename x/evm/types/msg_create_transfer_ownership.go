package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCreateTransferOwnershipRequest is the constructor for CreateTransferOwnershipRequest
func NewCreateTransferOwnershipRequest(sender sdk.AccAddress, chain string, keyID string) *CreateTransferOwnershipRequest {
	return &CreateTransferOwnershipRequest{Sender: sender, Chain: chain, KeyID: keyID}
}

// Route implements sdk.Msg
func (m CreateTransferOwnershipRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m CreateTransferOwnershipRequest) Type() string {
	return "CreateTransferOwnership"
}

// GetSignBytes  implements sdk.Msg
func (m CreateTransferOwnershipRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m CreateTransferOwnershipRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m CreateTransferOwnershipRequest) ValidateBasic() error {
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
