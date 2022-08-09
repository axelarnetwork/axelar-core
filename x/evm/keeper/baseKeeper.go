package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var (
	chainPrefix    = utils.KeyFromStr("chain")
	subspacePrefix = utils.KeyFromStr("subspace")
)

var _ types.BaseKeeper = BaseKeeper{}

// BaseKeeper implements the base Keeper
type BaseKeeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	// It is not safe to access subspaces directly (subspaces cannot be deleted so a subspace might exist for a chain that was deleted).
	// Use getSubspace to access a subspace.
	paramsKeeper types.ParamsKeeper
	subspaces    map[string]params.Subspace
}

// NewKeeper returns a new EVM base keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramsKeeper types.ParamsKeeper) BaseKeeper {
	return BaseKeeper{
		cdc:          cdc,
		storeKey:     storeKey,
		paramsKeeper: paramsKeeper,
		subspaces:    make(map[string]params.Subspace),
	}
}

// Logger returns a module-specific logger.
func (k BaseKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ForChain returns the keeper associated to the given chain
func (k BaseKeeper) ForChain(chain nexus.ChainName) types.ChainKeeper {
	return chainKeeper{
		BaseKeeper:    k,
		chainLowerKey: strings.ToLower(chain.String()),
	}
}

func (k BaseKeeper) getBaseStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k BaseKeeper) getStore(ctx sdk.Context, chain string) utils.KVStore {
	pre := string(chainPrefix.Append(utils.LowerCaseKey(chain)).AsKey()) + "_"
	return utils.NewNormalizedStore(prefix.NewStore(ctx.KVStore(k.storeKey), []byte(pre)), k.cdc)
}

// HasChain returns true if the chain has been set up
func (k BaseKeeper) HasChain(ctx sdk.Context, chain nexus.ChainName) bool {
	return k.getBaseStore(ctx).Has(subspacePrefix.AppendStr(strings.ToLower(chain.String())))
}
