package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeregisterChainMaintainerRequest creates a message of type DeregisterChainMaintainerRequest
func NewDeregisterChainMaintainerRequest(sender sdk.AccAddress, chains ...string) *DeregisterChainMaintainerRequest {
	var normalizedChains []string
	for _, chain := range chains {
		normalizedChains = append(normalizedChains, utils.NormalizeString(chain))
	}

	return &DeregisterChainMaintainerRequest{
		Sender: sender,
		Chains: normalizedChains,
	}
}

// Route implements sdk.Msg
func (m DeregisterChainMaintainerRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m DeregisterChainMaintainerRequest) Type() string {
	return "DeregisterChainMaintainer"
}

// ValidateBasic implements sdk.Msg
func (m DeregisterChainMaintainerRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if len(m.Chains) == 0 {
		return fmt.Errorf("missing chains")
	}

	for _, chain := range m.Chains {
		if err := utils.ValidateString(chain); err != nil {
			return sdkerrors.Wrap(err, "invalid chain")
		}
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m DeregisterChainMaintainerRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m DeregisterChainMaintainerRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
