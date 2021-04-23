package exported

import (
	"fmt"
)

// VotingData is needed so that the amino codec can (un)marshal the voting data correctly
type VotingData interface {
}

// NewPollMeta constructor for PollMeta without nonce
func NewPollMeta(module string, id string) PollMeta {
	return PollMeta{
		Module: module,
		ID:     id,
		Nonce:  0,
	}
}

// NewPollMetaWithNonce constructor for PollMeta with nonce; PollMeta with nonce can be re-voted
func NewPollMetaWithNonce(module string, id string, blockHeight int64, lockingPeriod int64) PollMeta {
	return PollMeta{
		Module: module,
		ID:     id,
		Nonce:  blockHeight / lockingPeriod,
	}
}

func (m PollMeta) String() string {
	return fmt.Sprintf("%s_%s_%d", m.Module, m.ID, m.Nonce)
}

// Validate performs a stateless validity check to ensure PollMeta has been properly initialized
func (m PollMeta) Validate() error {
	if m.Module == "" {
		return fmt.Errorf("missing module")
	}

	if m.ID == "" {
		return fmt.Errorf("missing poll ID")
	}

	return nil
}
