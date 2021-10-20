package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterChainMaintainerRequest creates a message of type RegisterChainMaintainerRequest
func NewRegisterChainMaintainerRequest(sender sdk.AccAddress, chains ...string) *RegisterChainMaintainerRequest {
	return &RegisterChainMaintainerRequest{
		Sender: sender,
		Chains: chains,
	}
}

// Route implements sdk.Msg
func (m RegisterChainMaintainerRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m RegisterChainMaintainerRequest) Type() string {
	return "RegisterChainMaintainer"
}

// ValidateBasic implements sdk.Msg
func (m RegisterChainMaintainerRequest) ValidateBasic() error {
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
func (m RegisterChainMaintainerRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m RegisterChainMaintainerRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
