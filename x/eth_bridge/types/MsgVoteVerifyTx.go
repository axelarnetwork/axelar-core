package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ exported.MsgVote = &MsgVoteVerifiedTx{}

type MsgVoteVerifiedTx struct {
	Sender     sdk.AccAddress
	PollMeta   exported.PollMeta
	VotingData bool
}

func (msg MsgVoteVerifiedTx) Poll() exported.PollMeta {
	return msg.PollMeta
}

func (msg *MsgVoteVerifiedTx) Data() exported.VotingData {
	// dummy return, must not be empty, otherwise marshaling crashes
	return msg.VotingData
}

func (msg MsgVoteVerifiedTx) Route() string {
	return RouterKey
}

func (msg MsgVoteVerifiedTx) Type() string {
	return "MsgVoteVerifiedTx"
}

func (msg MsgVoteVerifiedTx) ValidateBasic() error {
	if msg.Sender == nil {
		return fmt.Errorf("missing sender")
	}
	return msg.PollMeta.Validate()
}

func (msg MsgVoteVerifiedTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVoteVerifiedTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg *MsgVoteVerifiedTx) SetSender(address sdk.AccAddress) {
	msg.Sender = address
}
