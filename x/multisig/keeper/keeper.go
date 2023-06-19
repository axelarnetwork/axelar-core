package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
)

var (
	keygenPrefix           = utils.KeyFromInt(1)
	signingPrefix          = utils.KeyFromInt(2)
	keyPrefix              = utils.KeyFromInt(3)
	expiryKeygenPrefix     = utils.KeyFromInt(4)
	expirySigningPrefix    = utils.KeyFromInt(5)
	keyEpochPrefix         = utils.KeyFromInt(6)
	keyRotationCountPrefix = utils.KeyFromInt(7)
	signingSessionCountKey = utils.KeyFromInt(100)

	keygenOptOutPrefix = key.RegisterStaticKey(types.ModuleName, 8)
)

var _ types.Keeper = &Keeper{}

// Keeper provides access to all state changes regarding this module
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace
	sigRouter  types.SigRouter
}

// NewKeeper is the constructor for the keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace.WithKeyTable(types.KeyTable()),
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

func (k Keeper) setParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
