package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRegisterAssetRequest is the constructor for RegisterAssetRequest
func NewRegisterAssetRequest(sender sdk.AccAddress, chain string, asset nexus.Asset, isNativeAsset bool) *RegisterAssetRequest {
	return &RegisterAssetRequest{
		Sender:        sender,
		Chain:         utils.NormalizeString(chain),
		Asset:         asset,
		IsNativeAsset: isNativeAsset,
	}
}

// Route returns the route for this message
func (m RegisterAssetRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m RegisterAssetRequest) Type() string {
	return "RegisterAsset"
}

// ValidateBasic executes a stateless message validation
func (m RegisterAssetRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := m.Asset.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m RegisterAssetRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m RegisterAssetRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
