package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// MsgConfirmERC20Deposit represents an erc20 deposit confirmation message
type MsgConfirmERC20Deposit struct {
	Sender     sdk.AccAddress
	TxID       string
	Amount     sdk.Uint
	BurnerAddr string
}

// NewMsgConfirmERC20Deposit creates a message of type MsgConfirmERC20Deposit
func NewMsgConfirmERC20Deposit(sender sdk.AccAddress, txID common.Hash, amount sdk.Uint, burnerAddr common.Address) sdk.Msg {

	return MsgConfirmERC20Deposit{
		Sender:     sender,
		TxID:       txID.Hex(),
		Amount:     amount,
		BurnerAddr: burnerAddr.Hex(),
	}
}

// Route implements sdk.Msg
func (msg MsgConfirmERC20Deposit) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgConfirmERC20Deposit) Type() string {
	return "ConfirmERC20Deposit"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConfirmERC20Deposit) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgConfirmERC20Deposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgConfirmERC20Deposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
