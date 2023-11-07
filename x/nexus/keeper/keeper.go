package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
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

	chainMaintainerStatePrefix = key.RegisterStaticKey(types.ModuleName, 1)
	rateLimitPrefix            = key.RegisterStaticKey(types.ModuleName, 2)
	transferEpochPrefix        = key.RegisterStaticKey(types.ModuleName, 3)
	generalMessagePrefix       = key.RegisterStaticKey(types.ModuleName, 4)
	processingMessagePrefix    = key.RegisterStaticKey(types.ModuleName, 5)
	messageNonceKey            = key.RegisterStaticKey(types.ModuleName, 6)

	// temporary
	// TODO: add description about what temporary means
	latestDepositAddressPrefix = utils.KeyFromStr("latest_deposit_address")
)

// Keeper represents a nexus keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	params   params.Subspace

	addressValidator types.AddressValidator
	messageRouter    types.MessageRouter
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

// SetAddressValidator sets the nexus address validator. It will panic if called more than once
func (k *Keeper) SetAddressValidator(validator types.AddressValidator) {
	if k.addressValidator != nil {
		panic("validator already set")
	}

	k.addressValidator = validator

	// In order to avoid invalid or non-deterministic behavior, we seal the validator immediately
	// to prevent additionals handlers from being registered after the keeper is initialized.
	k.addressValidator.Seal()
}

// getAddressValidator returns the nexus address validator. If not set, it returns a sealed empty validator
func (k Keeper) getAddressValidator() types.AddressValidator {
	if k.addressValidator == nil {
		k.SetAddressValidator(types.NewAddressValidator())
	}

	return k.addressValidator
}

func (k *Keeper) SetMessageRouter(router types.MessageRouter) {
	if k.messageRouter != nil {
		panic("router already set")
	}

	k.messageRouter = router
	// In order to avoid invalid or non-deterministic behavior, we seal the router immediately
	// to prevent additionals handlers from being registered after the keeper is initialized.
	k.messageRouter.Seal()
}

func (k Keeper) getMessageRouter() types.MessageRouter {
	if k.messageRouter == nil {
		k.SetMessageRouter(types.NewMessageRouter())
	}

	return k.messageRouter
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
