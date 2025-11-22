package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// NewConfirmDepositRequest creates a message of type ConfirmDepositRequest
func NewConfirmDepositRequest(sender sdk.AccAddress, denom string, depositAddr sdk.AccAddress) *ConfirmDepositRequest {
	return &ConfirmDepositRequest{
		Sender:         sender.String(),
		Denom:          utils.NormalizeString(denom),
		DepositAddress: depositAddr,
	}
}

// Route implements sdk.Msg
func (m ConfirmDepositRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmDepositRequest) Type() string {
	return "ConfirmDeposit"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmDepositRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := sdk.ValidateDenom(m.Denom); err != nil {
		return errorsmod.Wrap(err, "invalid token denomination")
	}

	if err := sdk.VerifyAddressFormat(m.DepositAddress); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "deposit address").Error())
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmDepositRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
