package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewAddCosmosBasedChainRequest is the constructor for NewAddCosmosBasedChainRequest
func NewAddCosmosBasedChainRequest(sender sdk.AccAddress, name, addrPrefix string, assets []nexus.Asset) *AddCosmosBasedChainRequest {
	return &AddCosmosBasedChainRequest{
		Sender:       sender,
		AddrPrefix:   utils.NormalizeString(addrPrefix),
		NativeAssets: assets,
		CosmosChain:  nexus.ChainName(utils.NormalizeString(name)),
	}
}

// Route returns the route for this message
func (m AddCosmosBasedChainRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m AddCosmosBasedChainRequest) Type() string {
	return "AddCosmosBasedChain"
}

// ValidateBasic executes a stateless message validation
func (m AddCosmosBasedChainRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.AddrPrefix); err != nil {
		return sdkerrors.Wrap(err, "invalid address prefix")
	}

	seen := make(map[string]bool)
	for _, asset := range m.NativeAssets {
		if err := asset.Validate(); err != nil {
			return sdkerrors.Wrap(err, "invalid asset")
		}

		if !asset.IsNativeAsset {
			return fmt.Errorf("%s is not specified as a native asset", asset.Denom)
		}

		if seen[asset.Denom] {
			return fmt.Errorf("duplicate asset %s", asset.Denom)
		}

		seen[asset.Denom] = true
	}

	if err := m.CosmosChain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid cosmos chain name")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m AddCosmosBasedChainRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m AddCosmosBasedChainRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
