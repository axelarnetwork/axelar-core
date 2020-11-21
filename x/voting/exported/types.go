package exported

import (
	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

// this interface is needed so that the amino codec can (un)marshal the voting data correctly
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
	// Data returns the data that was voted on. Modules need to ensure they cast it back into the correct type
	Data() VotingData
}

type MsgVote interface {
	exported.MsgWithSenderSetter
	Vote
}
