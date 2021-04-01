package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// MsgConfirmDeposit represents an erc20 deposit confirmation message
type MsgConfirmDeposit struct {
	Sender     sdk.AccAddress
	TxID       string
	Amount     sdk.Uint
	BurnerAddr string
}

// NewMsgConfirmERC20Deposit creates a message of type MsgConfirmDeposit
func NewMsgConfirmERC20Deposit(sender sdk.AccAddress, txID common.Hash, amount sdk.Uint, burnerAddr common.Address) sdk.Msg {

	return MsgConfirmDeposit{
		Sender:     sender,
		TxID:       txID.Hex(),
		Amount:     amount,
		BurnerAddr: burnerAddr.Hex(),
	}
}

// Route implements sdk.Msg
func (msg MsgConfirmDeposit) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgConfirmDeposit) Type() string {
	return "ConfirmERC20Deposit"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConfirmDeposit) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgConfirmDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgConfirmDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
