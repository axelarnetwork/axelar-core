package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewCreateTransferOwnershipRequest is the constructor for CreateTransferOwnershipRequest
func NewCreateTransferOwnershipRequest(sender sdk.AccAddress, chain string, keyID string) *CreateTransferOwnershipRequest {
	return &CreateTransferOwnershipRequest{Sender: sender, Chain: utils.NormalizeString(chain), KeyID: tss.KeyID(keyID)}
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

	if err := utils.ValidateString(m.Chain, utils.DefaultDelimiter); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := m.KeyID.Validate(); err != nil {
		return err
	}

	return nil
}
