package keeper

import (
	"crypto/sha256"
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

// SetPendingRefund saves pending refundable message
func (k Keeper) SetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest, fee sdk.Coin) error {
	hash := sha256.Sum256(k.cdc.MustMarshalBinaryLengthPrefixed(&req))
	k.getStore(ctx).Set(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])), &fee)
	return nil
}

// GetPendingRefund retrieves a pending refundable message
func (k Keeper) GetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest) (sdk.Coin, bool) {
	var fee sdk.Coin
	hash := sha256.Sum256(k.cdc.MustMarshalBinaryLengthPrefixed(&req))
	ok := k.getStore(ctx).Get(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])), &fee)

	return fee, ok
}

// DeletePendingRefund retrieves a pending refundable message
func (k Keeper) DeletePendingRefund(ctx sdk.Context, req types.RefundMsgRequest) {
	hash := sha256.Sum256(k.cdc.MustMarshalBinaryLengthPrefixed(&req))
	k.getStore(ctx).Delete(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])))
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
