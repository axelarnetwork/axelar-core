package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// NewDeactivateChainRequest creates a message of type DeactivateChainRequest
func NewDeactivateChainRequest(sender sdk.AccAddress, chains ...string) *DeactivateChainRequest {
	var normalizedChains []string
	for _, chain := range chains {
		normalizedChains = append(normalizedChains, utils.NormalizeString(chain))
	}

	return &DeactivateChainRequest{
		Sender: sender,
		Chains: normalizedChains,
	}
}

// Route implements sdk.Msg
func (m DeactivateChainRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m DeactivateChainRequest) Type() string {
	return "DeactivateChain"
}

// ValidateBasic implements sdk.Msg
func (m DeactivateChainRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if len(m.Chains) == 0 {
		return fmt.Errorf("missing chains")
	}

	for _, chain := range m.Chains {
		if err := utils.ValidateString(chain); err != nil {
			return sdkerrors.Wrap(err, fmt.Sprintf("invalid chain '%s'", chain))
		}
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m DeactivateChainRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m DeactivateChainRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
