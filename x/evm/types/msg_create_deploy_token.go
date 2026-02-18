package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewCreateDeployTokenRequest is the constructor for CreateDeployTokenRequest
func NewCreateDeployTokenRequest(sender sdk.AccAddress, chain string, asset Asset, tokenDetails TokenDetails, address Address, dailyMintLimit string) *CreateDeployTokenRequest {
	return &CreateDeployTokenRequest{
		Sender:         sender.String(),
		Chain:          nexus.ChainName(utils.NormalizeString(chain)),
		Asset:          asset,
		TokenDetails:   tokenDetails,
		Address:        address,
		DailyMintLimit: dailyMintLimit,
	}
}

// Route implements sdk.Msg
func (m CreateDeployTokenRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m CreateDeployTokenRequest) Type() string {
	return "CreateDeployToken"
}

// GetSignBytes  implements sdk.Msg
func (m CreateDeployTokenRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m CreateDeployTokenRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	if err := m.Asset.Validate(); err != nil {
		return err
	}

	switch m.Address.IsZeroAddress() {
	case true:
		if m.Chain.Equals(m.Asset.Chain) {
			return fmt.Errorf("cannot deploy token on the origin chain")
		}
	case false:
		if !m.Chain.Equals(m.Asset.Chain) {
			return fmt.Errorf("cannot link token on a different chain")
		}
	}

	if err := m.TokenDetails.Validate(); err != nil {
		return err
	}

	// DailyMintLimit is deprecated and ignored - no validation needed

	return nil
}
