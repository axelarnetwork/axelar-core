package vald

import (
	"encoding/json"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

//go:generate moq -pkg mock -out ./mock/state.go . ReadWriter

// ReadWriter represents a data source/sink
type ReadWriter interface {
	WriteAll([]byte) error
	ReadAll() ([]byte, error)
}

// StateStore manages event state persistence
type StateStore struct {
	rw ReadWriter
}

// NewStateStore returns a new StateStore instance
func NewStateStore(rw ReadWriter) StateStore {
	return StateStore{rw: rw}
}

// GetState returns the stored block height for which all events have been published
func (s StateStore) GetState() (completed int64, err error) {
	bz, err := s.rw.ReadAll()
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not read the event state")
	}

	err = json.Unmarshal(bz, &completed)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "state is in unexpected format")
	}

	if completed < 0 {
		return 0, fmt.Errorf("state must be a positive integer")
	}

	return completed, nil
}

// SetState persists the block height for which all events have been published
func (s StateStore) SetState(completed int64) error {
	if completed < 0 {
		return fmt.Errorf("state must be a positive integer")
	}

	bz, err := json.Marshal(completed)
	if err != nil {
		return err
	}
	return s.rw.WriteAll(bz)
}
