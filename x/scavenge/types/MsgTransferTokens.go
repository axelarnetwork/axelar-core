package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure MsgTransferTokens implements sdk.Msg
var _ sdk.Msg = &MsgTransferTokens{}

type MsgTransferTokens struct {
	Sender    	sdk.AccAddress 	`json:"sender" yaml:"sender"`
	Recipient 	sdk.AccAddress 	`json:"recipient" yaml:"recipient"`
	Amount  	sdk.Coins	`json:"amount" yaml:"amount"`
}

// NewMsgRevealSolution creates a new MsgRevealSolution instance
func NewMsgTransferTokens(sender sdk.AccAddress, recipient sdk.AccAddress, amount sdk.Coins) MsgTransferTokens {

	return MsgTransferTokens{
		Sender:    sender,
		Recipient: recipient,
		Amount:     amount,
	}
}

// RevealSolutionConst is RevealSolution Constant
const TransferTokensConst = "TransferTokens"

// nolint
func (msg MsgTransferTokens) Route() string { return RouterKey }
func (msg MsgTransferTokens) Type() string  { return TransferTokensConst }
func (msg MsgTransferTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgTransferTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgTransferTokens) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender can't be empty")
	}
	if msg.Recipient.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient can't be empty")
	}
	if !msg.Amount.IsAllPositive()  {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "amount must be a positive integer")
	}
	return nil
}
