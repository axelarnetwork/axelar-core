package exported

import (
	"fmt"
)

// VotingData is needed so that the amino codec can (un)marshal the voting data correctly
type VotingData interface {
}

// PollMeta represents the meta data for a poll
type PollMeta struct {
	Module string
	Type   string
	ID     string
	Nonce  int64
}

// NewPollMeta constructor for PollMeta without nonce
func NewPollMeta(module string, pollType string, id string) PollMeta {
	return PollMeta{
		Module: module,
		Type:   pollType,
		ID:     id,
		Nonce:  0,
	}
}

// NewPollMetaWithNonce constructor for PollMeta with nonce; PollMeta with nonce can be re-voted
func NewPollMetaWithNonce(module string, pollType string, id string, blockHeight int64, lockingPeriod int64) PollMeta {
	return PollMeta{
		Module: module,
		Type:   pollType,
		ID:     id,
		Nonce:  blockHeight / lockingPeriod,
	}
}

func (p PollMeta) String() string {
	return fmt.Sprintf("%s_%s_%s_%d", p.Module, p.Type, p.ID, p.Nonce)
}

// Validate performs a stateless validity check to ensure PollMeta has been properly initialized
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

// Vote provides functionality to interact with a vote for a poll
type Vote interface {
	Poll() PollMeta
	// Data returns the data that was voted on. Modules need to ensure they cast it back into the correct type
	Data() VotingData
}

// MsgVote
// Deprecated
type MsgVote interface {
	Vote
}
