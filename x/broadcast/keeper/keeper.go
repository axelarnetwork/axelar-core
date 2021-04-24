package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/broadcast/types"
)

var _ broadcast.Broadcaster = Keeper{}

const (
	proxyCountKey = "proxyCount"
)

// Keeper - the broadcast keeper
type Keeper struct {
	staker   types.Staker
	storeKey sdk.StoreKey
	cdc      *codec.LegacyAmino
}

// NewKeeper constructs a broadcast keeper
func NewKeeper(
	cdc *codec.LegacyAmino,
	storeKey sdk.StoreKey,
	stakingKeeper types.Staker,
) (Keeper, error) {
	return Keeper{
		staker:   stakingKeeper,
		storeKey: storeKey,
		cdc:      cdc,
	}, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// RegisterProxy registers a proxy address for a given principal, which can broadcast messages in the principal's name
func (k Keeper) RegisterProxy(ctx sdk.Context, principal sdk.ValAddress, proxy sdk.AccAddress) error {
	val := k.staker.Validator(ctx, principal)
	if val == nil {
		return fmt.Errorf("validator %s is unknown", principal.String())
	}
	k.Logger(ctx).Debug("getting proxy count")
	count := k.getProxyCount(ctx)

	storedProxy := ctx.KVStore(k.storeKey).Get(principal)
	if storedProxy != nil {
		ctx.KVStore(k.storeKey).Delete(storedProxy)
		count--
	}
	k.Logger(ctx).Debug("setting proxy")
	ctx.KVStore(k.storeKey).Set(proxy, principal)
	// Creating a reverse lookup
	ctx.KVStore(k.storeKey).Set(principal, proxy)
	count++
	k.Logger(ctx).Debug("setting proxy count")
	k.setProxyCount(ctx, count)
	k.Logger(ctx).Debug("done")
	return nil
}

// GetPrincipal returns the proxy address for a given principal address. Returns nil if not set.
func (k Keeper) GetPrincipal(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
	if proxy == nil {
		return nil
	}
	return ctx.KVStore(k.storeKey).Get(proxy)
}

// GetProxy returns the proxy address for a given principal address. Returns nil if not set.
func (k Keeper) GetProxy(ctx sdk.Context, principal sdk.ValAddress) sdk.AccAddress {
	return ctx.KVStore(k.storeKey).Get(principal)
}

func (k Keeper) setProxyCount(ctx sdk.Context, count int) {
	k.Logger(ctx).Debug(fmt.Sprintf("number of known proxies: %v", count))
	ctx.KVStore(k.storeKey).Set([]byte(proxyCountKey), k.cdc.MustMarshalBinaryLengthPrefixed(count))
}

func (k Keeper) getProxyCount(ctx sdk.Context) int {
	bz := ctx.KVStore(k.storeKey).Get([]byte(proxyCountKey))
	if bz == nil {
		return 0
	}
	var count int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &count)
	return count
}
