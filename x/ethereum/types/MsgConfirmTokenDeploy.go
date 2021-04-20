package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewMsgConfirmERC20TokenDeploy creates a message of type MsgConfirmToken
func NewMsgConfirmERC20TokenDeploy(sender sdk.AccAddress, txID common.Hash, symbol string) *MsgConfirmToken {
	return &MsgConfirmToken{
		Sender: sender,
		TxID:   txID.Hex(),
		Symbol: symbol,
	}
}

// Route implements sdk.Msg
func (m MsgConfirmToken) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgConfirmToken) Type() string {
	return "ConfirmERC20TokenDeploy"
}

// ValidateBasic implements sdk.Msg
func (m MsgConfirmToken) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m MsgConfirmToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m MsgConfirmToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
