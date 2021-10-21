package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewDeregisterChainMaintainerRequest creates a message of type DeregisterChainMaintainerRequest
func NewDeregisterChainMaintainerRequest(sender sdk.AccAddress, chains ...string) *DeregisterChainMaintainerRequest {
	return &DeregisterChainMaintainerRequest{
		Sender: sender,
		Chains: chains,
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
		if chain == "" {
			return fmt.Errorf("chain cannot be empty")
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
