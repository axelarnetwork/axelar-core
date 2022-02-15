package legacy

import (
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/legacy/exported"
)

var (
	chainStatePrefix = utils.KeyFromStr("state")
	chainPrefix      = utils.KeyFromStr("chain")
)

// GetChains unmarshals Chain struct in v0.13
func GetChains(store utils.KVStore) (chains []exported.Chain) {
	iter := store.Iterator(chainPrefix)
	for ; iter.Valid(); iter.Next() {
		var chain exported.Chain
		iter.UnmarshalValue(&chain)
		chains = append(chains, chain)
	}

	return chains
}

// GetChainState unmarshals ChainState struct in v0.13
func GetChainState(store utils.KVStore, chain exported.Chain) (chainState ChainState, ok bool) {
	return chainState, store.Get(chainStatePrefix.Append(utils.LowerCaseKey(chain.Name)), &chainState)
}
