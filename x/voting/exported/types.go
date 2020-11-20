package exported

import (
	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

type VotingData interface {
}

type PollMeta struct {
	Module string
	Type   string
	ID     string
}

func (p PollMeta) String() string {
	return p.Module + p.Type + p.ID
}

type Vote interface {
	Poll() PollMeta
	Data() VotingData
	Confirms() bool
}

type MsgVote interface {
	exported.MsgWithProxySender
	Vote
}
