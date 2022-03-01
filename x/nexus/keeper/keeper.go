package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

var (
	nonceKey = utils.KeyFromStr("nonce")

	chainPrefix              = utils.KeyFromStr("chain")
	chainStatePrefix         = utils.KeyFromStr("state")
	chainByNativeAssetPrefix = utils.KeyFromStr("native_asset_chain")
	linkedAddressesPrefix    = utils.KeyFromStr("linked_addresses")
	transferPrefix           = utils.KeyFromStr("transfer")
	transferFee              = utils.KeyFromStr("fee")
	assetFeePrefix           = utils.KeyFromStr("asset_fee")
	// temporary
	latestDepositAddressPrefix = utils.KeyFromStr("latest_deposit_address")
)

// Keeper represents a nexus keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	params   params.Subspace
	router   types.Router
}

// NewKeeper returns a new nexus keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the nexus module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
}

// GetParams gets the nexus module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// SetRouter sets the nexus router. It will panic if called more than once
func (k *Keeper) SetRouter(router types.Router) {
	if k.router != nil {
		panic("router already set")
	}

	k.router = router

	// In order to avoid invalid or non-deterministic behavior, we seal the router immediately
	// to prevent additionals handlers from being registered after the keeper is initialized.
	k.router.Seal()
}

// GetRouter returns the nexus router. If no router was set, it returns a (sealed) router with no handlers
func (k Keeper) GetRouter() types.Router {
	if k.router == nil {
		k.SetRouter(types.NewRouter())
	}

	return k.router
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
