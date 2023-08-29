package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRegisterAssetRequest is the constructor for RegisterAssetRequest
func NewRegisterAssetRequest(sender sdk.AccAddress, chain string, asset nexus.Asset, limit sdk.Uint, window time.Duration) *RegisterAssetRequest {
	return &RegisterAssetRequest{
		Sender: sender,
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
		Asset:  asset,
		Limit:  limit,
		Window: window,
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

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := m.Asset.Validate(); err != nil {
		return err
	}

	// Any m.Limit value is valid. If m.Limit is equal to 0, it means no cross-chain transfers will be allowed.

	if m.Window.Nanoseconds() <= 0 {
		return fmt.Errorf("rate limit window must be positive")
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
