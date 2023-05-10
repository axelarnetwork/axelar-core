package evm

import (
	"sync"
)

//go:generate moq -out ./mock/latest_finalized_block_cache.go -pkg mock . LatestFinalizedBlockCache

// LatestFinalizedBlockCache is a cache for the latest finalized block number for each chain
type LatestFinalizedBlockCache interface {
	// Get returns the latest finalized block number for chain
	Get(chain string) uint64
	// Set sets the latest finalized block number for chain, if the given block number is greater than the current latest finalized block number
	Set(chain string, blockNumber uint64)
}

type latestFinalizedBlockCache struct {
	cache map[string]uint64
	lock  sync.RWMutex
}

func NewLatestFinalizedBlockCache() LatestFinalizedBlockCache {
	return &latestFinalizedBlockCache{
		cache: make(map[string]uint64),
		lock:  sync.RWMutex{},
	}
}

// Get returns the latest finalized block number for chain
func (c *latestFinalizedBlockCache) Get(chain string) uint64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.cache[chain]
}

// Set sets the latest finalized block number for chain, if the given block number is greater than the current latest finalized block number
func (c *latestFinalizedBlockCache) Set(chain string, blockNumber uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if blockNumber > c.cache[chain] {
		c.cache[chain] = blockNumber
	}
}
