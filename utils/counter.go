package utils

import (
	"encoding/binary"

	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/utils/convert"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"golang.org/x/exp/constraints"
)

type Counter[T constraints.Unsigned] interface {
	// Incr increments the counter and returns the value before the increment
	Incr(ctx sdk.Context) T
	// Curr returns the current value of the counter
	Curr(ctx sdk.Context) T
}

type counter[T constraints.Unsigned] struct {
	key    key.Key
	store  KVStore
	logger log.Logger
}

func NewCounter[T constraints.Unsigned](key key.Key, store KVStore, logger log.Logger) Counter[T] {
	return counter[T]{
		key:    key,
		store:  store,
		logger: logger,
	}
}

func (c counter[T]) Incr(ctx sdk.Context) T {
	curr := c.Curr(ctx)
	defer c.store.SetRawNew(c.key, convert.IntToBytes(curr+1))

	return curr
}

func (c counter[T]) Curr(ctx sdk.Context) T {
	return T(binary.BigEndian.Uint64(c.store.GetRawNew(c.key)))
}
