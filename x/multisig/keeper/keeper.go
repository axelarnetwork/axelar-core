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
	keygenPrefix           = utils.KeyFromInt(1)   // Deprecated: migrate to keygenPrefixNew
	signingPrefix          = utils.KeyFromInt(2)   // Deprecated: migrate to signingPrefixNew
	keyPrefix              = utils.KeyFromInt(3)   // Deprecated: migrate to keyPrefixNew
	expiryKeygenPrefix     = utils.KeyFromInt(4)   // Deprecated: migrate to expiryKeygenPrefixNew
	expirySigningPrefix    = utils.KeyFromInt(5)   // Deprecated: migrate to expirySigningPrefixNew
	keyEpochPrefix         = utils.KeyFromInt(6)   // Deprecated: migrate to keyEpochPrefixNew
	keyRotationCountPrefix = utils.KeyFromInt(7)   // Deprecated: migrate to keyRotationCountPrefixNew
	signingSessionCountKey = utils.KeyFromInt(100) // Deprecated: migrate to signingSessionCountKeyNew

	keygenPrefixNew           = key.RegisterStaticKey(types.ModuleName, 1)
	signingPrefixNew          = key.RegisterStaticKey(types.ModuleName, 2)
	keyPrefixNew              = key.RegisterStaticKey(types.ModuleName, 3)
	expiryKeygenPrefixNew     = key.RegisterStaticKey(types.ModuleName, 4)
	expirySigningPrefixNew    = key.RegisterStaticKey(types.ModuleName, 5)
	keyEpochPrefixNew         = key.RegisterStaticKey(types.ModuleName, 6)
	keyRotationCountPrefixNew = key.RegisterStaticKey(types.ModuleName, 7)
	keygenOptOutPrefix        = key.RegisterStaticKey(types.ModuleName, 8)
	signingSessionCountKeyNew = key.RegisterStaticKey(types.ModuleName, 100)
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

func (k Keeper) getParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

func (k Keeper) setParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
