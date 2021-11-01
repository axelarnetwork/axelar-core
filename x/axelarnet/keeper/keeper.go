package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

var (
	pathPrefix          = utils.KeyFromStr("path")
	pendingRefundPrefix = utils.KeyFromStr("refund")
	cosmosChainPrefix   = utils.KeyFromStr("cosmos_chain")
	ibcAssetPrefix      = utils.KeyFromStr("ibc_asset")
	feeCollector        = utils.KeyFromStr("fee_collector")
)

// Keeper provides access to all state changes regarding the Axelarnet module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	params   params.Subspace
}

// NewKeeper returns a new axelarnet keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams gets the axelarnet module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

// SetParams sets the axelarnet module's parameters
func (k Keeper) SetParams(ctx sdk.Context, n types.Nexus, p types.Params) {
	k.params.SetParamSet(ctx, &p)
	for _, c := range p.SupportedChains {
		chain, ok := n.GetChain(ctx, c)
		if ok {
			n.RegisterAsset(ctx, exported.Axelarnet.Name, chain.NativeAsset)
		}
	}
}

// GetRouteTimeoutWindow returns the timeout window for IBC transfers routed by axelarnet
func (k Keeper) GetRouteTimeoutWindow(ctx sdk.Context) uint64 {
	var result uint64
	k.params.Get(ctx, types.KeyRouteTimeoutWindow, &result)

	return result
}

// RegisterIBCPath registers an IBC path for a cosmos chain
func (k Keeper) RegisterIBCPath(ctx sdk.Context, chain, path string) error {
	bz := k.getStore(ctx).GetRaw(pathPrefix.Append(utils.LowerCaseKey(chain)))
	if bz != nil {
		return fmt.Errorf("chain %s already registered", chain)
	}
	k.getStore(ctx).SetRaw(pathPrefix.Append(utils.LowerCaseKey(chain)), []byte(path))
	return nil
}

// SetPendingRefund saves pending refundable message
func (k Keeper) SetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest, fee sdk.Coin) error {
	hash := sha256.Sum256(k.cdc.MustMarshalLengthPrefixed(&req))
	k.getStore(ctx).Set(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])), &fee)
	return nil
}

// GetPendingRefund retrieves a pending refundable message
func (k Keeper) GetPendingRefund(ctx sdk.Context, req types.RefundMsgRequest) (sdk.Coin, bool) {
	var fee sdk.Coin
	hash := sha256.Sum256(k.cdc.MustMarshalLengthPrefixed(&req))
	ok := k.getStore(ctx).Get(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])), &fee)

	return fee, ok
}

// DeletePendingRefund retrieves a pending refundable message
func (k Keeper) DeletePendingRefund(ctx sdk.Context, req types.RefundMsgRequest) {
	hash := sha256.Sum256(k.cdc.MustMarshalLengthPrefixed(&req))
	k.getStore(ctx).Delete(pendingRefundPrefix.Append(utils.KeyFromBz(hash[:])))
}

// GetIBCPath retrieves the IBC path associated to the specified chain
func (k Keeper) GetIBCPath(ctx sdk.Context, chain string) (string, bool) {
	bz := k.getStore(ctx).GetRaw(pathPrefix.Append(utils.LowerCaseKey(chain)))
	if bz == nil {
		return "", false
	}

	return string(bz), true
}

// SetPendingIBCTransfer saves a pending IBC transfer routed by axelarnet
func (k Keeper) SetPendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64, value types.IBCTransfer) {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, sequence)
	key := pathPrefix.Append(utils.KeyFromStr(portID)).Append(utils.KeyFromStr(channelID)).Append(utils.KeyFromBz(bz))

	k.getStore(ctx).Set(key, &value)
}

// GetPendingIBCTransfer gets a pending IBC transfer routed by axelarnet
func (k Keeper) GetPendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64) (types.IBCTransfer, bool) {
	var value types.IBCTransfer
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, sequence)
	key := pathPrefix.Append(utils.KeyFromStr(portID)).Append(utils.KeyFromStr(channelID)).Append(utils.KeyFromBz(bz))

	ok := k.getStore(ctx).Get(key, &value)
	return value, ok
}

// DeletePendingIBCTransfer deletes a pending IBC transfer routed by axelarnet
func (k Keeper) DeletePendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64) {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, sequence)
	key := pathPrefix.Append(utils.KeyFromStr(portID)).Append(utils.KeyFromStr(channelID)).Append(utils.KeyFromBz(bz))

	k.getStore(ctx).Delete(key)
}

// GetCosmosChains retrieves all registered cosmos chains
func (k Keeper) GetCosmosChains(ctx sdk.Context) []string {
	var results []string
	iter := k.getStore(ctx).Iterator(cosmosChainPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		results = append(results, string(bz))
	}

	return results
}

// RegisterAssetToCosmosChain sets an asset origins from a cosmos chain
func (k Keeper) RegisterAssetToCosmosChain(ctx sdk.Context, asset string, chain string) {
	store := k.getStore(ctx)
	// register a cosmos chain to axelarnet
	key := cosmosChainPrefix.Append(utils.LowerCaseKey(chain))
	if !store.Has(key) {
		k.getStore(ctx).SetRaw(key, []byte(strings.ToLower(chain)))
	}
	// register asset to the cosmos chain
	k.getStore(ctx).SetRaw(ibcAssetPrefix.Append(utils.LowerCaseKey(asset)), []byte(strings.ToLower(chain)))
}

// GetCosmosChain gets a asset's original chain
func (k Keeper) GetCosmosChain(ctx sdk.Context, asset string) (string, bool) {
	bz := k.getStore(ctx).GetRaw(ibcAssetPrefix.Append(utils.LowerCaseKey(asset)))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// SetFeeCollector sets axelarnet fee collector
func (k Keeper) SetFeeCollector(ctx sdk.Context, address sdk.AccAddress) {
	k.getStore(ctx).SetRaw(feeCollector, address)
}

// GetFeeCollector gets axelarnet fee collector
func (k Keeper) GetFeeCollector(ctx sdk.Context) (sdk.AccAddress, bool) {
	bz := k.getStore(ctx).GetRaw(feeCollector)
	if bz == nil {
		return sdk.AccAddress{}, false
	}

	return bz, true
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
