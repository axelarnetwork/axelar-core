package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

var _ exported.ValidatorMsg = &MsgBallot{}

type MsgBallot struct {
	Votes  []bool
	Sender sdk.AccAddress
}

func (msg *MsgBallot) SetSender(sender sdk.AccAddress) {
	msg.Sender = sender
}

func (msg MsgBallot) Route() string {
	return RouterKey
}

func (msg MsgBallot) Type() string {
	return "BatchVote"
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

func NewMsgBatchVote(votes []bool) *MsgBallot {
	return &MsgBallot{Votes: votes}
}
