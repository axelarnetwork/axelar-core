package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewLinkRequest creates a message of type LinkRequest
func NewLinkRequest(sender sdk.AccAddress, chain, recipientChain, recipientAddr, asset string) *LinkRequest {
	return &LinkRequest{
		Sender:         sender,
		Chain:          chain,
		RecipientChain: recipientChain,
		RecipientAddr:  recipientAddr,
		Asset:          asset,
	}
}

// Route implements sdk.Msg
func (m LinkRequest) Route() string {
	return RouterKey
}

// Type  implements sdk.Msg
func (m LinkRequest) Type() string {
	return "Link"
}

// ValidateBasic implements sdk.Msg
func (m LinkRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}
	if m.RecipientAddr == "" {
		return fmt.Errorf("missing recipient address")
	}
	if m.RecipientChain == "" {
		return fmt.Errorf("missing recipient chain")
	}

	if m.Asset == "" {
		return fmt.Errorf("missing asset name")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m LinkRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m LinkRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
