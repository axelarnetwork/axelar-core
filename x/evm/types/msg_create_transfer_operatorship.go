package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewCreateTransferOperatorshipRequest creates a message of type CreateTransferOperatorshipRequest
func NewCreateTransferOperatorshipRequest(sender sdk.AccAddress, chain string, keyID string) *CreateTransferOperatorshipRequest {
	return &CreateTransferOperatorshipRequest{
		Sender: sender,
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
		KeyID:  multisig.KeyID(keyID),
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

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := m.KeyID.ValidateBasic(); err != nil {
		return err
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
