package types

import (
	"fmt"
	"strings"

	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewCreateDeployTokenRequest is the constructor for CreateDeployTokenRequest
func NewCreateDeployTokenRequest(sender sdk.AccAddress, chain string, asset Asset, tokenDetails TokenDetails, minAmount sdk.Int, address Address) *CreateDeployTokenRequest {
	return &CreateDeployTokenRequest{
		Sender:       sender,
		Chain:        utils.NormalizeString(chain),
		Asset:        asset,
		TokenDetails: tokenDetails,
		MinAmount:    minAmount,
		Address:      address,
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

// GetSigners implements sdk.Msg
func (m CreateDeployTokenRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m CreateDeployTokenRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := m.Asset.Validate(); err != nil {
		return err
	}

	switch m.Address.IsZeroAddress() {
	case true:
		if strings.EqualFold(m.Chain, m.Asset.Chain) {
			return fmt.Errorf("cannot deploy token on the origin chain")
		}
	case false:
		if !strings.EqualFold(m.Chain, m.Asset.Chain) {
			return fmt.Errorf("cannot link token on a different chain")
		}
	}

	if err := m.TokenDetails.Validate(); err != nil {
		return err
	}

	if m.MinAmount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("minimum amount for mint/withdrawals must be greater than zero")
	}

	return nil
}
