package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewMsgConfirmERC20Deposit creates a message of type MsgConfirmDeposit
func NewMsgConfirmERC20Deposit(sender sdk.AccAddress, txID common.Hash, amount sdk.Uint, burnerAddr common.Address) *MsgConfirmDeposit {

	return &MsgConfirmDeposit{
		Sender:        sender,
		TxID:          txID.Hex(),
		Amount:        amount,
		BurnerAddress: burnerAddr.Hex(),
	}
}

// Route implements sdk.Msg
func (m MsgConfirmDeposit) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgConfirmDeposit) Type() string {
	return "ConfirmERC20Deposit"
}

// ValidateBasic implements sdk.Msg
func (m MsgConfirmDeposit) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m MsgConfirmDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m MsgConfirmDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
