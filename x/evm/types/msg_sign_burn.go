package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewSignBurnTokensRequest is the constructor for SignBurnTokensRequest
func NewSignBurnTokensRequest(sender sdk.AccAddress, chain string) *SignBurnTokensRequest {
	return &SignBurnTokensRequest{Sender: sender, Chain: chain}
}

// Route implements sdk.Msg
func (m SignBurnTokensRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SignBurnTokensRequest) Type() string {
	return "SignBurnTokens"
}

// GetSignBytes  implements sdk.Msg
func (m SignBurnTokensRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m SignBurnTokensRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m SignBurnTokensRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	return nil
}
