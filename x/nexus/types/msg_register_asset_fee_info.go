package types

import (
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewRegisterAssetFeeInfoRequest creates a message of type RegisterAssetFeeInfoRequest
func NewRegisterAssetFeeInfoRequest(sender sdk.AccAddress, chain string, asset string, feeInfo exported.FeeInfo) *RegisterAssetFeeInfoRequest {
	chain = utils.NormalizeString(chain)
	asset = utils.NormalizeString(asset)

	return &RegisterAssetFeeInfoRequest{
		Sender:  sender,
		Chain:   chain,
		Asset:   asset,
		FeeInfo: feeInfo,
	}
}

// Route implements sdk.Msg
func (m RegisterAssetFeeInfoRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m RegisterAssetFeeInfoRequest) Type() string {
	return "RegisterAssetFeeInfo"
}

// ValidateBasic implements sdk.Msg
func (m RegisterAssetFeeInfoRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := utils.ValidateString(m.Asset); err != nil {
		return sdkerrors.Wrap(err, "invalid asset")
	}

	if err := m.FeeInfo.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m RegisterAssetFeeInfoRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m RegisterAssetFeeInfoRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
