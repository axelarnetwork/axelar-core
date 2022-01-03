package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewActivateChainRequest creates a message of type ActivateChainRequest
func NewActivateChainRequest(sender sdk.AccAddress, chains ...string) *ActivateChainRequest {
	return &ActivateChainRequest{
		Sender: sender,
		Chains: chains,
	}
}

// Route implements sdk.Msg
func (m ActivateChainRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ActivateChainRequest) Type() string {
	return "ActivateChain"
}

// ValidateBasic implements sdk.Msg
func (m ActivateChainRequest) ValidateBasic() error {
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
func (m ActivateChainRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ActivateChainRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
