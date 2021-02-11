package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
)

// MsgLink represents the message that links a cross chain address to a burner address
type MsgLink struct {
	Sender      sdk.AccAddress
	Recipient   exported.CrossChainAddress
	Symbol      string
	GatewayAddr string
}

// NewMsgLink implements sdk.Msg
func NewMsgLink(sender sdk.AccAddress, destination exported.CrossChainAddress, symbol, gateway string) sdk.Msg {
	return MsgLink{
		Sender:      sender,
		Recipient:   destination,
		Symbol:      symbol,
		GatewayAddr: gateway,
	}
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
