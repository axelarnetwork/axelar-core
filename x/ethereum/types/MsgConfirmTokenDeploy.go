package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// MsgConfirmERC20TokenDeploy represents a token deploy confirmation message
type MsgConfirmERC20TokenDeploy struct {
	Sender sdk.AccAddress
	TxID   string
	Symbol string
}

// NewMsgConfirmERC20TokenDeploy creates a message of type MsgConfirmERC20TokenDeploy
func NewMsgConfirmERC20TokenDeploy(sender sdk.AccAddress, txID common.Hash, symbol string) sdk.Msg {
	return MsgConfirmERC20TokenDeploy{
		Sender: sender,
		TxID:   txID.Hex(),
		Symbol: symbol,
	}
}

// Route implements sdk.Msg
func (msg MsgConfirmERC20TokenDeploy) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgConfirmERC20TokenDeploy) Type() string {
	return "ConfirmERC20TokenDeploy"
}

// ValidateBasic implements sdk.Msg
func (msg MsgConfirmERC20TokenDeploy) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgConfirmERC20TokenDeploy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgConfirmERC20TokenDeploy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
