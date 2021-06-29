// Package testutils provides general purpose utility functions for unit/integration testing.
package testutils

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	cdc *codec.LegacyAmino
)

// Func wraps a regular testing function so it can be used as a pointer function receiver
type Func func(t *testing.T)

// Repeat executes the testing function n times
func (f Func) Repeat(n int) Func {
	return func(t *testing.T) {
		for i := 0; i < n; i++ {
			f(t)
		}
	}
}

// Events wraps sdk.Events
type Events []abci.Event

// Filter returns a collection of events filtered by the predicate
func (fe Events) Filter(predicate func(events abci.Event) bool) Events {
	var filtered Events
	for _, event := range fe {
		if predicate(event) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
