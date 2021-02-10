package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ exported.MsgVote = &MsgVoteVerifiedToken{}

type MsgVoteVerifiedToken struct {
	Sender     sdk.AccAddress
	PollMeta   exported.PollMeta
	VotingData bool
}

func (msg MsgVoteVerifiedToken) Poll() exported.PollMeta {
	return msg.PollMeta
}

func (msg *MsgVoteVerifiedToken) Data() exported.VotingData {
	return msg.VotingData
}

func (msg MsgVoteVerifiedToken) Route() string {
	return RouterKey
}

func (msg MsgVoteVerifiedToken) Type() string {
	return "VoteVerifiedToken"
}

func (msg MsgVoteVerifiedToken) ValidateBasic() error {
	if msg.Sender == nil {
		return fmt.Errorf("missing sender")
	}
	return msg.PollMeta.Validate()
}

func (msg MsgVoteVerifiedToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVoteVerifiedToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg *MsgVoteVerifiedToken) SetSender(address sdk.AccAddress) {
	msg.Sender = address
}
