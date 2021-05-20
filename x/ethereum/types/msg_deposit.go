package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewConfirmDepositRequest creates a message of type ConfirmDepositRequest
func NewConfirmDepositRequest(sender sdk.AccAddress, txID common.Hash, amount sdk.Uint, burnerAddr common.Address) *ConfirmDepositRequest {
	return &ConfirmDepositRequest{
		Sender:     sender,
		TxID:       txID.Hex(),
		Amount:     amount,
		BurnerAddr: burnerAddr.Hex(),
	}
}

// Route implements sdk.Msg
func (m ConfirmDepositRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmDepositRequest) Type() string {
	return "ConfirmERC20Deposit"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmDepositRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
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
