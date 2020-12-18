package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ exported.MsgVote = &MsgVoteMock{}

// MsgVoteMock is a mock implementation of exported.MsgVote.
type MsgVoteMock struct {
	RouteVal  string
	TypeVal   string
	SignBytes []byte
	Sender    sdk.AccAddress
	PollVal   exported.PollMeta
	DataVal   exported.VotingData
}

func (msg MsgVoteMock) Route() string {
	return msg.RouteVal
}

func (msg MsgVoteMock) Type() string {
	return msg.TypeVal
}

func (msg MsgVoteMock) ValidateBasic() error {
	return nil
}

func (msg MsgVoteMock) GetSignBytes() []byte {
	return msg.SignBytes
}

func (msg MsgVoteMock) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg *MsgVoteMock) SetSender(address sdk.AccAddress) {
	msg.Sender = address
}

func (msg MsgVoteMock) Poll() exported.PollMeta {
	return msg.PollVal
}

func (msg MsgVoteMock) Data() exported.VotingData {
	return msg.DataVal
}
