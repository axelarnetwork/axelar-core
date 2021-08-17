package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCreateBurnTokensRequest is the constructor for CreateBurnTokensRequest
func NewCreateBurnTokensRequest(sender sdk.AccAddress, chain string) *CreateBurnTokensRequest {
	return &CreateBurnTokensRequest{Sender: sender, Chain: chain}
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
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	return nil
}
