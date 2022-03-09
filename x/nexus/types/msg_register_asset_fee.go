package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRegisterAssetFeeRequest creates a message of type RegisterAssetFeeRequest
func NewRegisterAssetFeeRequest(sender sdk.AccAddress, feeInfo exported.FeeInfo) *RegisterAssetFeeRequest {
	return &RegisterAssetFeeRequest{
		Sender:  sender,
		FeeInfo: feeInfo,
	}
}

// Route implements sdk.Msg
func (m RegisterAssetFeeRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m RegisterAssetFeeRequest) Type() string {
	return "RegisterAssetFee"
}

// ValidateBasic implements sdk.Msg
func (m RegisterAssetFeeRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.FeeInfo.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m RegisterAssetFeeRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m RegisterAssetFeeRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
