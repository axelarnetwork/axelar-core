package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var (
	chainPrefix    = utils.KeyFromStr("chain")
	subspacePrefix = "subspace"
)

var _ types.BaseKeeper = &BaseKeeper{}

// BaseKeeper implements the base Keeper
type BaseKeeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	paramsKeeper types.ParamsKeeper
	initialized  bool
}

// NewKeeper returns a new EVM base keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramsKeeper types.ParamsKeeper) *BaseKeeper {
	return &BaseKeeper{
		cdc:          cdc,
		storeKey:     storeKey,
		paramsKeeper: paramsKeeper,
	}
}

// Logger returns a module-specific logger.
func (k BaseKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// InitChains initializes all existing EVM chains and their respective param subspaces
func (k *BaseKeeper) InitChains(ctx sdk.Context) {
	if k.initialized {
		panic("chains are already initialized")
	}

	iter := k.getBaseStore(ctx).IteratorNew(key.FromStr(subspacePrefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		_ = k.getSubspace(ctx, nexus.ChainName(iter.Key()))
	}

	k.initialized = true
}

// CreateChain creates the subspace for a new EVM chain. Returns an error if the chain already exists
func (k BaseKeeper) CreateChain(ctx sdk.Context, params types.Params) error {
	if !k.initialized {
		panic("InitChain must be called before chain keepers can be used")
	}

	chainKey := key.FromStr(subspacePrefix).Append(key.FromStr(params.Chain.String()))
	if k.getBaseStore(ctx).HasNew(chainKey) {
		return fmt.Errorf("chain %s already exists", params.Chain)
	}

	k.getBaseStore(ctx).SetRawNew(chainKey, []byte(params.Chain))

	subspace := k.getSubspace(ctx, params.Chain)
	subspace.SetParamSet(ctx, &params)
	return nil
}

// ForChain returns the keeper associated to the given chain
func (k BaseKeeper) ForChain(ctx sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
	if !k.initialized {
		panic("InitChain must be called before chain keepers can be used")
	}

	chainKey := key.FromStr(subspacePrefix).Append(key.FromStr(chain.String()))
	if !k.getBaseStore(ctx).HasNew(chainKey) {
		return chainKeeper{}, fmt.Errorf("unknown chain %s", chain)
	}

	return chainKeeper{
		BaseKeeper: k,
		chain:      chain,
	}, nil
}

func (k BaseKeeper) getBaseStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k BaseKeeper) getSubspace(ctx sdk.Context, chain nexus.ChainName) params.Subspace {
	chainKey := types.ModuleName + "_" + strings.ToLower(chain.String())
	if subspace, ok := k.paramsKeeper.GetSubspace(chainKey); ok {
		return subspace
	}

	k.Logger(ctx).Debug(fmt.Sprintf("initialized evm subspace %s", chain))
	return k.paramsKeeper.Subspace(chainKey).WithKeyTable(types.KeyTable())
}
