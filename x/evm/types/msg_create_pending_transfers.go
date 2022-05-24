package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewCreatePendingTransfersRequest - CreatePendingTransfersRequest constructor
func NewCreatePendingTransfersRequest(sender sdk.AccAddress, chain string) *CreatePendingTransfersRequest {
	return &CreatePendingTransfersRequest{Sender: sender, Chain: nexus.ChainName(utils.NormalizeString(chain))}
}

// Route returns the route for this message
func (m CreatePendingTransfersRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m CreatePendingTransfersRequest) Type() string {
	return "CreatePendingTransfers"
}

// ValidateBasic executes a stateless message validation
func (m CreatePendingTransfersRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m CreatePendingTransfersRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m CreatePendingTransfersRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
