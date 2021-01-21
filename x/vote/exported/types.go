package exported

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

// VotingData is needed so that the amino codec can (un)marshal the voting data correctly
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

func (p PollMeta) Validate() error {
	if p.Module == "" {
		return fmt.Errorf("missing module")
	}
	if p.Type == "" {
		return fmt.Errorf("missing poll type")
	}
	if p.ID == "" {
		return fmt.Errorf("missing poll ID")
	}
	return nil
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

// Voter is the interface that provides voting functionality to other modules
type Voter interface {
	InitPoll(ctx sdk.Context, poll PollMeta) error
	DeletePoll(ctx sdk.Context, poll PollMeta)
	RecordVote(ctx sdk.Context, vote MsgVote) error
	TallyVote(ctx sdk.Context, vote MsgVote) error
	Result(ctx sdk.Context, poll PollMeta) VotingData
}
