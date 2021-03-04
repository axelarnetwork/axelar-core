package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// MsgVerifyErc20TokenDeploy represents a token deploy verification message
type MsgVerifyErc20TokenDeploy struct {
	Sender sdk.AccAddress
	TxID   []byte
	Symbol string
}

// NewMsgVerifyErc20TokenDeploy creates a message of type MsgVerifyErc20TokenDeploy
func NewMsgVerifyErc20TokenDeploy(sender sdk.AccAddress, txID common.Hash, symbol string) sdk.Msg {
	return MsgVerifyErc20TokenDeploy{
		Sender: sender,
		TxID:   txID.Bytes(),
		Symbol: symbol,
	}
}

// Route implements sdk.Msg
func (msg MsgVerifyErc20TokenDeploy) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgVerifyErc20TokenDeploy) Type() string {
	return "VerifyErc20TokenDeploy"
}

// ValidateBasic implements sdk.Msg
func (msg MsgVerifyErc20TokenDeploy) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgVerifyErc20TokenDeploy) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgVerifyErc20TokenDeploy) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
