package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRegisterAssetRequest is the constructor for RegisterAssetRequest
func NewRegisterAssetRequest(sender sdk.AccAddress, chain string, asset nexus.Asset) *RegisterAssetRequest {
	return &RegisterAssetRequest{
		Sender: sender.String(),
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
		Asset:  asset,
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
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
