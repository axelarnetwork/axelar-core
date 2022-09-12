package utils

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/constraints"

	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/utils/convert"
)

// Counter is a stateful counter that works with the kv store and starts from zero
type Counter[T constraints.Unsigned] struct {
	key   key.Key
	store KVStore
}

// NewCounter is the constructor for counter
func NewCounter[T constraints.Unsigned](key key.Key, store KVStore) Counter[T] {
	return Counter[T]{
		key:   key,
		store: store,
	}
}

// Incr increments the counter and returns the value before the increment
func (c Counter[T]) Incr(ctx sdk.Context) T {
	curr := c.Curr(ctx)
	c.store.SetRawNew(c.key, convert.IntToBytes(curr+1))

	return curr
}

// Curr returns the current value of the counter
func (c Counter[T]) Curr(ctx sdk.Context) T {
	bz := c.store.GetRawNew(c.key)
	if bz == nil {
		return 0
	}

	return T(binary.BigEndian.Uint64(bz))
}
