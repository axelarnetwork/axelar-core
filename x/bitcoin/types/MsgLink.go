package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgLink represents a message to link an cross-chain address to a Bitcoin address
type MsgLink struct {
	Sender         sdk.AccAddress
	RecipientAddr  string
	RecipientChain string
}

// NewMsgLink - MsgLink constructor
func NewMsgLink(sender sdk.AccAddress, recipientAddr string, recipientChain string) sdk.Msg {
	return MsgLink{
		Sender:         sender,
		RecipientAddr:  recipientAddr,
		RecipientChain: recipientChain,
	}
}

// Route returns the route for this message
func (msg MsgLink) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgLink) Type() string {
	return "Link"
}

// ValidateBasic executes a stateless message validation
func (msg MsgLink) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	if msg.RecipientAddr == "" {
		return fmt.Errorf("missing recipient address")
	}
	if msg.RecipientChain == "" {
		return fmt.Errorf("missing recipient chain")
	}
	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgLink) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgLink) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
