package types

import (
	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewLinkRequest creates a message of type LinkRequest
func NewLinkRequest(sender sdk.AccAddress, chain, recipientChain, recipientAddr, asset string) *LinkRequest {
	return &LinkRequest{
		Sender:         sender,
		Chain:          utils.NormalizeString(chain),
		RecipientChain: utils.NormalizeString(recipientChain),
		RecipientAddr:  utils.NormalizeString(recipientAddr),
		Asset:          utils.NormalizeString(asset),
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
	if err := utils.ValidateString(m.Chain, utils.DefaultDelimiter); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}
	if err := utils.ValidateString(m.RecipientChain, utils.DefaultDelimiter); err != nil {
		return sdkerrors.Wrap(err, "invalid recipient chain")
	}
	if err := utils.ValidateString(m.RecipientAddr, utils.DefaultDelimiter); err != nil {
		return sdkerrors.Wrap(err, "invalid recipient address")
	}
	if err := utils.ValidateString(m.Asset, utils.DefaultDelimiter); err != nil {
		return sdkerrors.Wrap(err, "invalid asset")
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
