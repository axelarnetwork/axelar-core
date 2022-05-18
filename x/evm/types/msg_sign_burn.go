package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewCreateBurnTokensRequest is the constructor for CreateBurnTokensRequest
func NewCreateBurnTokensRequest(sender sdk.AccAddress, chain string) *CreateBurnTokensRequest {
	return &CreateBurnTokensRequest{Sender: sender, Chain: nexus.ChainName(utils.NormalizeString(chain))}
}

// Route implements sdk.Msg
func (m CreateBurnTokensRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m CreateBurnTokensRequest) Type() string {
	return "CreateBurnTokens"
}

// GetSignBytes  implements sdk.Msg
func (m CreateBurnTokensRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m CreateBurnTokensRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m CreateBurnTokensRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	return nil
}
