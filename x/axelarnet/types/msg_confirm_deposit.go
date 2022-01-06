package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewConfirmDepositRequest creates a message of type ConfirmDepositRequest
func NewConfirmDepositRequest(sender sdk.AccAddress, txID []byte, denom string, depositAddr sdk.AccAddress) *ConfirmDepositRequest {
	return &ConfirmDepositRequest{
		Sender:         sender,
		TxID:           txID,
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
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if len(m.TxID) != 32 {
		return fmt.Errorf("invalid TxID")
	}

	if err := utils.ValidateString(m.Denom); err != nil {
		return sdkerrors.Wrap(err, "invalid token denomination")
	}

	if err := sdk.VerifyAddressFormat(m.DepositAddress); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "deposit address").Error())
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmDepositRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmDepositRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
