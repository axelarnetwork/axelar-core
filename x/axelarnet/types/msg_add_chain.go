package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewAddCosmosBasedChainRequest is the constructor for NewAddCosmosBasedChainRequest
func NewAddCosmosBasedChainRequest(sender sdk.AccAddress, name, nativeAsset, addrPrefix string, minAmount sdk.Int) *AddCosmosBasedChainRequest {
	return &AddCosmosBasedChainRequest{
		Sender: sender,
		Chain: nexus.Chain{
			Name:                  utils.NormalizeString(name),
			NativeAsset:           utils.NormalizeString(nativeAsset),
			SupportsForeignAssets: true,
			KeyType:               tss.None,
			Module:                "axelarnet", // cannot use constant due to import cycle
		},
		AddrPrefix: utils.NormalizeString(addrPrefix),
		MinAmount:  minAmount,
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

	if err := m.Chain.Validate(); err != nil {
		return fmt.Errorf("invalid chain spec: %v", err)
	}

	if err := utils.ValidateString(m.AddrPrefix); err != nil {
		return sdkerrors.Wrap(err, "invalid address prefix")
	}

	if m.MinAmount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("minimum mint/withdrawal amount must be greater than zero")
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
