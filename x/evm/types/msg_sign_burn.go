package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewCreateBurnTokensRequest is the constructor for CreateBurnTokensRequest
func NewCreateBurnTokensRequest(sender sdk.AccAddress, chain string) *CreateBurnTokensRequest {
	return &CreateBurnTokensRequest{Sender: sender.String(), Chain: nexus.ChainName(utils.NormalizeString(chain))}
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

// ValidateBasic implements sdk.Msg
func (m CreateBurnTokensRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	return nil
}
