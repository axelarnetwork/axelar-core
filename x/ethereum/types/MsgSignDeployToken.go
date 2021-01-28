package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSignDeployToken represents the message to sign a deploy token command for AxelarGateway
type MsgSignDeployToken struct {
	Sender    sdk.AccAddress
	Capacity  sdk.Int
	Decimals  uint8
	Symbol    string
	TokenName string
}

// Route implements sdk.Msg
func (msg MsgSignDeployToken) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgSignDeployToken) Type() string {
	return "SignDeployToken"
}

// GetSignBytes  implements sdk.Msg
func (msg MsgSignDeployToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgSignDeployToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// ValidateBasic implements sdk.Msg
func (msg MsgSignDeployToken) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.TokenName == "" {
		return fmt.Errorf("missing token name")
	}
	if msg.Symbol == "" {
		return fmt.Errorf("missing token symbol")
	}
	if !msg.Capacity.IsPositive() {
		return fmt.Errorf("token capacity must be a positive number")
	}
	return nil
}

// NewMsgSignDeployToken is the constructor for MsgSignDeployToken
func NewMsgSignDeployToken(sender sdk.AccAddress, tokenName string, symbol string, decimals uint8, capacity sdk.Int) sdk.Msg {
	return MsgSignDeployToken{Sender: sender, TokenName: tokenName, Symbol: symbol, Decimals: decimals, Capacity: capacity}
}
