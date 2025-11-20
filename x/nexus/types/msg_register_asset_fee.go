package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewRegisterAssetFeeRequest creates a message of type RegisterAssetFeeRequest
func NewRegisterAssetFeeRequest(sender sdk.AccAddress, feeInfo exported.FeeInfo) *RegisterAssetFeeRequest {
	return &RegisterAssetFeeRequest{
		Sender:  sender.String(),
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
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
