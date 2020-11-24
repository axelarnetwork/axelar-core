package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	brExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

var _ brExported.MsgWithSenderSetter = &MsgBallot{}

// MsgBallot holds subjective validator opinions and is used to broadcast them to the network
type MsgBallot struct {
	// each vote represents vote by the local validator regarding a different currently open poll
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

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (msg MsgBallot) ValidateBasic() error {
	// the individual votes' ValidateBasic function is not called here because we should not discard a whole ballot
	// because of a single faulty vote
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
