package keeper

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/cosmos/cosmos-sdk/codec"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	pendingChainKey = utils.KeyFromStr("pending_chain_asset")

	chainPrefix    = utils.KeyFromStr("chain")
	subspacePrefix = utils.KeyFromStr("subspace")
)

var _ types.BaseKeeper = baseKeeper{}

// Keeper implements both the base chainKeeper and chain chainKeeper
type baseKeeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	// It is not safe to access subspaces directly (subspaces cannot be deleted so a subspace might exist for a chain that was deleted).
	// Use getSubspace to access a subspace.
	paramsKeeper types.ParamsKeeper
	subspaces    map[string]params.Subspace
}

// NewKeeper returns a new EVM base keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramsKeeper types.ParamsKeeper) types.BaseKeeper {
	return baseKeeper{
		cdc:          cdc,
		storeKey:     storeKey,
		paramsKeeper: paramsKeeper,
		subspaces:    make(map[string]params.Subspace),
	}
}

// Logger returns a module-specific logger.
func (k baseKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ForChain returns the keeper associated to the given chain
func (k baseKeeper) ForChain(chain string) types.ChainKeeper {
	return chainKeeper{
		baseKeeper:    k,
		chainLowerKey: strings.ToLower(chain),
	}
}

// SetPendingChain stores the chain pending for confirmation
func (k baseKeeper) SetPendingChain(ctx sdk.Context, chain nexus.Chain, p types.Params) {
	k.getBaseStore(ctx).Set(pendingChainKey.Append(utils.LowerCaseKey(chain.Name)), &types.PendingChain{Chain: chain, Params: p})
}

// GetPendingChain returns the chain object with the given name, false if the chain is either unknown or confirmed
func (k baseKeeper) GetPendingChain(ctx sdk.Context, chainName string) (types.PendingChain, bool) {
	var chain types.PendingChain
	found := k.getBaseStore(ctx).Get(pendingChainKey.Append(utils.LowerCaseKey(chainName)), &chain)

	return chain, found
}

// DeletePendingChain deletes a chain that is not registered yet
func (k baseKeeper) DeletePendingChain(ctx sdk.Context, chain string) {
	k.getBaseStore(ctx).Delete(pendingChainKey.Append(utils.LowerCaseKey(chain)))
}

func (k baseKeeper) getBaseStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k baseKeeper) getStore(ctx sdk.Context, chain string) utils.KVStore {
	pre := string(chainPrefix.Append(utils.LowerCaseKey(chain)).AsKey()) + "_"
	return utils.NewNormalizedStore(prefix.NewStore(ctx.KVStore(k.storeKey), []byte(pre)), k.cdc)
}
