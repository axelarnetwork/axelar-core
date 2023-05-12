package evm

import (
	"math/big"
	"strings"
	"sync"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

//go:generate moq -out ./mock/latest_finalized_block_cache.go -pkg mock . LatestFinalizedBlockCache

// LatestFinalizedBlockCache is a cache for the latest finalized block number for each chain
type LatestFinalizedBlockCache interface {
	// Get returns the latest finalized block number for chain
	Get(chain nexus.ChainName) *big.Int
	// Set sets the latest finalized block number for chain, if the given block number is greater than the current latest finalized block number
	Set(chain nexus.ChainName, blockNumber *big.Int)
}

type latestFinalizedBlockCache struct {
	cache map[string]*big.Int
	lock  sync.RWMutex
}

func NewLatestFinalizedBlockCache() LatestFinalizedBlockCache {
	return &latestFinalizedBlockCache{
		cache: make(map[string]*big.Int),
		lock:  sync.RWMutex{},
	}
}

// Get returns the latest finalized block number for chain
func (c *latestFinalizedBlockCache) Get(chain nexus.ChainName) *big.Int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cachedBlockNumber, ok := c.cache[strings.ToLower(chain.String())]
	if !ok {
		return big.NewInt(0)
	}

	return cachedBlockNumber
}

// Set sets the latest finalized block number for chain, if the given block number is greater than the current latest finalized block number
func (c *latestFinalizedBlockCache) Set(chain nexus.ChainName, blockNumber *big.Int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	chainName := strings.ToLower(chain.String())

	cachedBlockNumber, ok := c.cache[chainName]
	if !ok || blockNumber.Cmp(cachedBlockNumber) > 0 {
		c.cache[chainName] = blockNumber
	}
}
