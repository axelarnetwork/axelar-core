package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// MsgVerifyErc20Deposit represents an erc20 deposit verification message
type MsgVerifyErc20Deposit struct {
	Sender     sdk.AccAddress
	TxID       [common.HashLength]byte
	Amount     sdk.Uint
	BurnerAddr string
}

// NewMsgVerifyErc20Deposit creates a message of type MsgVerifyErc20Deposit
func NewMsgVerifyErc20Deposit(sender sdk.AccAddress, txID common.Hash, amount sdk.Uint, burnerAddr common.Address) sdk.Msg {
	var array [common.HashLength]byte
	copy(array[:], txID.Bytes())

	return MsgVerifyErc20Deposit{
		Sender:     sender,
		TxID:       array,
		Amount:     amount,
		BurnerAddr: burnerAddr.Hex(),
	}
}

// Route implements sdk.Msg
func (msg MsgVerifyErc20Deposit) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgVerifyErc20Deposit) Type() string {
	return "VerifyErc20Deposit"
}

// ValidateBasic implements sdk.Msg
func (msg MsgVerifyErc20Deposit) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgVerifyErc20Deposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgVerifyErc20Deposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
