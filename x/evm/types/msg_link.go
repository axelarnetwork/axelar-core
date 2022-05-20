package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewLinkRequest creates a message of type LinkRequest
func NewLinkRequest(sender sdk.AccAddress, chain, recipientChain, recipientAddr, asset string) *LinkRequest {
	return &LinkRequest{
		Sender:         sender,
		Chain:          nexus.ChainName(utils.NormalizeString(chain)),
		RecipientChain: nexus.ChainName(utils.NormalizeString(recipientChain)),
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

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := m.RecipientChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid recipient chain")
	}

	if err := utils.ValidateString(m.RecipientAddr); err != nil {
		return sdkerrors.Wrap(err, "invalid recipient address")
	}

	if err := sdk.ValidateDenom(m.Asset); err != nil {
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
