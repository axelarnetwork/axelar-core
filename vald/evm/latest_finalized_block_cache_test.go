package evm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/vald/evm"
)

func TestGetSet(t *testing.T) {
	cache := evm.NewLatestFinalizedBlockCache()

	cache.Set("chain1", 1)
	cache.Set("chain2", 10)
	assert.Equal(t, cache.Get("chain1"), uint64(1))
	assert.Equal(t, cache.Get("chain2"), uint64(10))

	cache.Set("chain1", 2)
	assert.Equal(t, cache.Get("chain1"), uint64(2))

	cache.Set("chain1", 1)
	assert.Equal(t, cache.Get("chain1"), uint64(2))
}
