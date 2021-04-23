package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Route implements sdk.Msg
func (m MsgSignDeployToken) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgSignDeployToken) Type() string {
	return "SignDeployToken"
}

// GetSignBytes  implements sdk.Msg
func (m MsgSignDeployToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m MsgSignDeployToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m MsgSignDeployToken) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.TokenName == "" {
		return fmt.Errorf("missing token name")
	}
	if m.Symbol == "" {
		return fmt.Errorf("missing token symbol")
	}
	if !m.Capacity.IsPositive() {
		return fmt.Errorf("token capacity must be a positive number")
	}
	return nil
}

// NewMsgSignDeployToken is the constructor for MsgSignDeployToken
func NewMsgSignDeployToken(sender sdk.AccAddress, tokenName string, symbol string, decimals uint8, capacity sdk.Int) *MsgSignDeployToken {
	return &MsgSignDeployToken{
		Sender:    sender,
		TokenName: tokenName,
		Symbol:    symbol,
		Decimals:  decimals,
		Capacity:  capacity,
	}
}
