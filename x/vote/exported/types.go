package exported

import (
	"fmt"
)

// NewPollKey constructor for PollKey without nonce
func NewPollKey(module string, id string) PollKey {
	return PollKey{
		Module: module,
		ID:     id,
	}
}

func (m PollKey) String() string {
	return fmt.Sprintf("%s_%s", m.Module, m.ID)
}

// Validate performs a stateless validity check to ensure PollKey has been properly initialized
func (m PollKey) Validate() error {
	if m.Module == "" {
		return fmt.Errorf("missing module")
	}

	if m.ID == "" {
		return fmt.Errorf("missing poll ID")
	}

	return nil
}
