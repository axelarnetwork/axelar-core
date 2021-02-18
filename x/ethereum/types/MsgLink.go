package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgLink represents the message that links a cross chain address to a burner address
type MsgLink struct {
	Sender         sdk.AccAddress
	RecipientAddr  string
	Symbol         string
	GatewayAddr    string
	RecipientChain string
}

// Route implements sdk.Msg
func (msg MsgLink) Route() string {
	return RouterKey
}

// Type  implements sdk.Msg
func (msg MsgLink) Type() string {
	return "Link"
}

// ValidateBasic implements sdk.Msg
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

	if msg.Symbol == "" {
		return fmt.Errorf("missing asset symbol")
	}

	if msg.GatewayAddr == "" {
		return fmt.Errorf("missing gateway address")
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgLink) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgLink) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
