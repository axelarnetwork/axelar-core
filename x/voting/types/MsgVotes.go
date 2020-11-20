package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	brExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

var _ brExported.MsgWithProxySender = &MsgBallot{}

type MsgBallot struct {
	Votes  []exported.MsgVote
	Sender sdk.AccAddress
}

func (msg *MsgBallot) SetSender(sender sdk.AccAddress) {
	msg.Sender = sender
}

func (msg MsgBallot) Route() string {
	return RouterKey
}

func (msg MsgBallot) Type() string {
	return "SendBallot"
}

func (msg MsgBallot) ValidateBasic() error {
	if msg.Votes == nil {
		return fmt.Errorf("votes must not be nil")
	}
	return nil
}

func (msg MsgBallot) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgBallot) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
