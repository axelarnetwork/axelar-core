package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewLinkRequest creates a message of type LinkRequest
func NewLinkRequest(sender sdk.AccAddress, recipientChain, recipientAddr, asset string) *LinkRequest {
	return &LinkRequest{
		Sender:         sender,
		RecipientAddr:  recipientAddr,
		RecipientChain: recipientChain,
		Asset:          asset,
	}
}

// Route returns the route for this message
func (m LinkRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m LinkRequest) Type() string {
	return "Link"
}

// ValidateBasic executes a stateless message validation
func (m LinkRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.RecipientChain, false); err != nil {
		return fmt.Errorf("invalid recipient chain: %s", err.Error())
	}

	if err := utils.ValidateString(m.RecipientAddr, false); err != nil {
		return fmt.Errorf("invalid recipient address: %s", err.Error())
	}

	if err := utils.ValidateString(m.Asset, false); err != nil {
		return fmt.Errorf("invalid asset: %s", err.Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m LinkRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m LinkRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
