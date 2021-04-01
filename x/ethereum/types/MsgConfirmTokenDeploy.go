package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// MsgConfirmToken represents a token deploy confirmation message
type MsgConfirmToken struct {
	Sender sdk.AccAddress
	TxID   string
	Symbol string
}

// NewMsgConfirmERC20TokenDeploy creates a message of type MsgConfirmToken
func NewMsgConfirmERC20TokenDeploy(sender sdk.AccAddress, txID common.Hash, symbol string) sdk.Msg {
	return MsgConfirmToken{
		Sender: sender,
		TxID:   txID.Hex(),
		Symbol: symbol,
	}
}

// Route implements sdk.Msg
func (msg MsgConfirmToken) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgConfirmToken) Type() string {
	return "ConfirmERC20TokenDeploy"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConfirmToken) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgConfirmToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgConfirmToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
