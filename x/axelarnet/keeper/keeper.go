package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

var (
	pathPrefix          = utils.KeyFromStr("path_")
	pendingRefundPrefix = utils.KeyFromStr("refund_")
)

// Keeper provides access to all state changes regarding the Axelarnet module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler
}

// NewKeeper returns a new nexus keeper
func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, ) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// RegisterIBCPath registers an IBC path for an asset
func (k Keeper) RegisterIBCPath(ctx sdk.Context, asset, path string) error {
	bz := k.getStore(ctx).GetRaw(pathPrefix.Append(utils.LowerCaseKey(asset)))
	if bz != nil {
		return fmt.Errorf("asset %s already registered", asset)
	}
	k.getStore(ctx).SetRaw(pathPrefix.Append(utils.LowerCaseKey(asset)), []byte(path))
	return nil
}

// SetPotentialRefund saves potential refundable transactions
func (k Keeper) SetPotentialRefund(ctx sdk.Context, msgHash []byte, amt sdk.Coin) error {
	bz := k.cdc.MustMarshalBinaryBare(&amt)
	k.getStore(ctx).SetRaw(pendingRefundPrefix.Append(utils.KeyFromBz(msgHash)), bz)
	return nil
}

// GetPotentialRefund retrieves a potential refundable transactions
func (k Keeper) GetPotentialRefund(ctx sdk.Context, msgHash []byte) (sdk.Coin, bool) {
	bz := k.getStore(ctx).GetRaw(pendingRefundPrefix.Append(utils.KeyFromBz(msgHash)))
	if bz == nil {
		return sdk.Coin{}, false
	}

	var fee sdk.Coin
	k.cdc.MustUnmarshalBinaryBare(bz, &fee)

	return fee, true
}

// GetIBCPath retrieves the IBC path associated to the specified asset
func (k Keeper) GetIBCPath(ctx sdk.Context, asset string) string {
	bz := k.getStore(ctx).GetRaw(pathPrefix.Append(utils.LowerCaseKey(asset)))
	if bz == nil {
		return ""
	}
	return string(bz)
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
