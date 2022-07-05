package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

var (
	keygenPrefix = utils.KeyFromStr("session_keygen")
	keyPrefix    = utils.KeyFromStr("key")
)

// Keeper provides access to all state changes regarding this module
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace
}

// NewKeeper is the constructor for the keeper
func NewKeeper(storeKey sdk.StoreKey, cdc codec.BinaryCodec, paramSpace paramtypes.Subspace) Keeper {
	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramSpace: paramSpace,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

// SetParams sets the parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// CreateKeygenSession creates a new keygen session with the given key ID and snapshot
func (k Keeper) CreateKeygenSession(ctx sdk.Context, id exported.KeyID, snapshot snapshot.Snapshot) error {
	if _, ok := k.GetKeygenSession(ctx, id); ok {
		return fmt.Errorf("key %s already being generated", id)
	}

	if _, ok := k.getKey(ctx, id); ok {
		return fmt.Errorf("key %s already set", id)
	}

	panic("TODO")
}

// GetKeygenSession returns the keygen session with the given key ID
func (k Keeper) GetKeygenSession(ctx sdk.Context, id exported.KeyID) (keygen types.KeygenSession, ok bool) {
	return keygen, k.getStore(ctx).Get(keygenPrefix.AppendStr(id.String()), &keygen)
}

// GetKey returns the key with the given key ID
func (k Keeper) GetKey(ctx sdk.Context, id exported.KeyID) (exported.Key, bool) {
	panic("TODO")
}

// DeleteKeygenSession deletes the keygen session with the given key ID
func (k Keeper) DeleteKeygenSession(ctx sdk.Context, id exported.KeyID) {
	k.getStore(ctx).Delete(keygenPrefix.AppendStr(id.String()))
}

// SetKey stores the given key
func (k Keeper) SetKey(ctx sdk.Context, key types.Key) {
	k.getStore(ctx).Set(keyPrefix.AppendStr(key.ID.String()), &key)
}

func (k Keeper) setKeygenSession(ctx sdk.Context, keygen types.KeygenSession) {
	k.getStore(ctx).Set(keygenPrefix.AppendStr(keygen.GetKeyID().String()), &keygen)
}

func (k Keeper) getKey(ctx sdk.Context, id exported.KeyID) (key types.Key, ok bool) {
	return key, k.getStore(ctx).Get(keyPrefix.AppendStr(id.String()), &key)
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
