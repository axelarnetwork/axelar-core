package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// NewCreateApproveContractCallsRequest - CreateApproveContractCallsRequest constructor
func NewCreateApproveContractCallsRequest(sender sdk.AccAddress, chain string) *CreateApproveContractCallsRequest {
	return &CreateApproveContractCallsRequest{Sender: sender, Chain: utils.NormalizeString(chain)}
}

// Route returns the route for this message
func (m CreateApproveContractCallsRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m CreateApproveContractCallsRequest) Type() string {
	return "CreateApproveContractCalls"
}

// ValidateBasic executes a stateless message validation
func (m CreateApproveContractCallsRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m CreateApproveContractCallsRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m CreateApproveContractCallsRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
