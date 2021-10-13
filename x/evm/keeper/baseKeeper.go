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
	storeKey     sdk.StoreKey
	cdc          codec.BinaryCodec
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
		baseKeeper: k,
		chain:      strings.ToLower(chain),
	}
}

// SetPendingChain stores the chain pending for confirmation
func (k baseKeeper) SetPendingChain(ctx sdk.Context, chain nexus.Chain) {
	k.getStore(ctx, chain.Name).Set(pendingChainKey, &chain)
}

// GetPendingChain returns the chain object with the given name, false if the chain is either unknown or confirmed
func (k baseKeeper) GetPendingChain(ctx sdk.Context, chainName string) (nexus.Chain, bool) {
	var chain nexus.Chain
	found := k.getStore(ctx, chainName).Get(pendingChainKey, &chain)

	return chain, found
}

// DeletePendingChain deletes a chain that is not registered yet
func (k baseKeeper) DeletePendingChain(ctx sdk.Context, chain string) {
	k.getStore(ctx, chain).Delete(pendingChainKey)
}

// SetParams sets the evm module's parameters
func (k baseKeeper) SetParams(ctx sdk.Context, params ...types.Params) {
	for _, p := range params {
		chain := strings.ToLower(p.Chain)

		// set the chain before calling the subspace so it is recognized as an existing chain
		k.getBaseStore(ctx).SetRaw(subspacePrefix.AppendStr(chain), []byte(p.Chain))
		subspace, _ := k.getSubspace(ctx, chain)
		subspace.SetParamSet(ctx, &p)
	}
}

// GetParams gets the evm module's parameters
func (k baseKeeper) GetParams(ctx sdk.Context) []types.Params {
	ps := make([]types.Params, 0)
	iter := k.getBaseStore(ctx).Iterator(subspacePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		chain := string(iter.Value())
		subspace, _ := k.getSubspace(ctx, chain)

		var p types.Params
		subspace.GetParamSet(ctx, &p)
		ps = append(ps, p)
	}

	return ps
}

func (k baseKeeper) getBaseStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k baseKeeper) getStore(ctx sdk.Context, chain string) utils.KVStore {
	pre := string(chainPrefix.Append(utils.LowerCaseKey(chain)).AsKey()) + "_"
	return utils.NewNormalizedStore(prefix.NewStore(ctx.KVStore(k.storeKey), []byte(pre)), k.cdc)
}

func (k baseKeeper) getSubspace(ctx sdk.Context, chain string) (params.Subspace, bool) {
	chainLower := strings.ToLower(chain)

	// When a node restarts or joins the network after genesis, it might not have all EVM subspaces initialized.
	// The following checks has to be done regardless, if we would only do it dependent on the existence of a subspace
	// different nodes would consume different amounts of gas and it would result in a consensus failure
	if !k.getBaseStore(ctx).Has(subspacePrefix.AppendStr(chainLower)) {
		return params.Subspace{}, false
	}

	chainKey := types.ModuleName + "_" + chainLower
	subspace, ok := k.subspaces[chainKey]
	if !ok {
		subspace = k.paramsKeeper.Subspace(chainKey).WithKeyTable(types.KeyTable())
		k.subspaces[chainKey] = subspace
	}
	return subspace, true
}
