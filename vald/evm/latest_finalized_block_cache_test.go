package evm_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/vald/evm"
)

func TestGetSet(t *testing.T) {
	cache := evm.NewLatestFinalizedBlockCache()

	cache.Set("chain1", big.NewInt(1))
	cache.Set("chain2", big.NewInt(10))
	assert.Equal(t, cache.Get("chain1"), big.NewInt(1))
	assert.Equal(t, cache.Get("chain2"), big.NewInt(10))

	cache.Set("chain1", big.NewInt(2))
	assert.Equal(t, cache.Get("chain1"), big.NewInt(2))

	cache.Set("chain1", big.NewInt(1))
	assert.Equal(t, cache.Get("chain1"), big.NewInt(2))
}
