package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

var (
	pathPrefix         = utils.KeyFromStr("path")
	transferPrefix     = utils.KeyFromStr("transfer")
	addrPrefixPrefix   = utils.KeyFromStr("addr_prefix")
	cosmosChainPrefix  = utils.KeyFromStr("cosmos_chain")
	chainByAssetPrefix = utils.KeyFromStr("chain_by_asset")
	assetByChainPrefix = utils.KeyFromStr("asset_by_chain")
	feeCollector       = utils.KeyFromStr("fee_collector")
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

func (k Keeper) getParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

func (k Keeper) setParams(ctx sdk.Context, n types.Nexus, p types.Params) {
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

// GetTransactionFeeRate returns the transaction fee rate for axelarnet and cosmos chains
func (k Keeper) GetTransactionFeeRate(ctx sdk.Context) sdk.Dec {
	var result sdk.Dec
	k.params.Get(ctx, types.KeyTransactionFeeRate, &result)

	return result
}

// RegisterIBCPath registers an IBC path for a cosmos chain
func (k Keeper) RegisterIBCPath(ctx sdk.Context, chain, path string) error {
	key := pathPrefix.Append(utils.LowerCaseKey(chain))

	if k.getStore(ctx).GetRaw(key) != nil {
		return fmt.Errorf("chain %s already registered", chain)
	}
	k.getStore(ctx).SetRaw(key, []byte(path))
	return nil
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
func (k Keeper) SetPendingIBCTransfer(ctx sdk.Context, transfer types.IBCTransfer) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, transfer.Sequence)
	key := transferPrefix.Append(utils.KeyFromStr(transfer.PortID)).Append(utils.KeyFromStr(transfer.ChannelID)).Append(utils.KeyFromBz(bz))

	k.getStore(ctx).Set(key, &transfer)
}

// GetPendingIBCTransfer gets a pending IBC transfer routed by axelarnet
func (k Keeper) GetPendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64) (types.IBCTransfer, bool) {
	var value types.IBCTransfer
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, sequence)
	key := transferPrefix.Append(utils.KeyFromStr(portID)).Append(utils.KeyFromStr(channelID)).Append(utils.KeyFromBz(bz))

	ok := k.getStore(ctx).Get(key, &value)
	return value, ok
}

func (k Keeper) getPendingIBCTransfers(ctx sdk.Context) []types.IBCTransfer {
	iter := k.getStore(ctx).Iterator(transferPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var transfers []types.IBCTransfer
	for ; iter.Valid(); iter.Next() {
		var transfer types.IBCTransfer
		iter.UnmarshalValue(&transfer)
		transfers = append(transfers, transfer)
	}

	return transfers
}

// DeletePendingIBCTransfer deletes a pending IBC transfer routed by axelarnet
func (k Keeper) DeletePendingIBCTransfer(ctx sdk.Context, portID, channelID string, sequence uint64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, sequence)
	key := pathPrefix.Append(utils.KeyFromStr(portID)).Append(utils.KeyFromStr(channelID)).Append(utils.KeyFromBz(bz))

	k.getStore(ctx).Delete(key)
}

// GetCosmosChainByAsset gets an asset's original chain
func (k Keeper) GetCosmosChainByAsset(ctx sdk.Context, asset string) (types.CosmosChain, bool) {
	bz := k.getStore(ctx).GetRaw(assetByChainPrefix.Append(utils.LowerCaseKey(asset)))
	if bz == nil {
		return types.CosmosChain{}, false
	}

	chain, ok := k.GetCosmosChainByName(ctx, string(bz))
	if !ok {
		return types.CosmosChain{}, false
	}

	return chain, true
}

// GetCosmosChains retrieves all registered cosmos chains
func (k Keeper) GetCosmosChains(ctx sdk.Context) []string {
	var results []string
	iter := k.getStore(ctx).Iterator(cosmosChainPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var value types.CosmosChain
		iter.UnmarshalValue(&value)
		results = append(results, value.Name)
	}

	return results
}

// RegisterAssetToCosmosChain sets an asset's original cosmos chain
func (k Keeper) RegisterAssetToCosmosChain(ctx sdk.Context, asset string, chain string) {
	if registeredChain, ok := k.GetCosmosChainByAsset(ctx, asset); ok && registeredChain.Name != chain {
		k.deleteAssetByChain(ctx, registeredChain.Name, asset)
	}

	k.setAssetByChain(ctx, chain, asset)
	k.setChainByAsset(ctx, asset, chain)
}

func (k Keeper) deleteAssetByChain(ctx sdk.Context, chain string, asset string) {
	k.getStore(ctx).Delete(assetByChainPrefix.
		Append(utils.LowerCaseKey(chain)).
		Append(utils.LowerCaseKey(asset)))
}

func (k Keeper) setAssetByChain(ctx sdk.Context, chain string, asset string) {
	k.getStore(ctx).SetRaw(assetByChainPrefix.
		Append(utils.LowerCaseKey(chain)).
		Append(utils.LowerCaseKey(asset)), []byte(asset))
}

func (k Keeper) setChainByAsset(ctx sdk.Context, asset string, chain string) {
	k.getStore(ctx).SetRaw(chainByAssetPrefix.Append(utils.LowerCaseKey(asset)), []byte(chain))
}

func (k Keeper) getAssets(ctx sdk.Context, chain string) []string {
	iter := k.getStore(ctx).Iterator(assetByChainPrefix.Append(utils.LowerCaseKey(chain)))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var assets []string
	for ; iter.Valid(); iter.Next() {
		assets = append(assets, string(iter.Value()))
	}

	return assets
}

// SetFeeCollector sets axelarnet fee collector
func (k Keeper) SetFeeCollector(ctx sdk.Context, address sdk.AccAddress) {
	if address != nil {
		k.getStore(ctx).SetRaw(feeCollector, address)
	}
}

// SetCosmosChain sets the address prefix for the given cosmos chain
func (k Keeper) SetCosmosChain(ctx sdk.Context, chain types.CosmosChain) {
	// register a cosmos chain to axelarnet
	key := cosmosChainPrefix.Append(utils.LowerCaseKey(chain.Name))
	if !k.getStore(ctx).Has(key) {
		k.getStore(ctx).Set(key, &chain)
	}
}

// GetFeeCollector gets axelarnet fee collector
func (k Keeper) GetFeeCollector(ctx sdk.Context) (sdk.AccAddress, bool) {
	bz := k.getStore(ctx).GetRaw(feeCollector)
	if bz == nil {
		return sdk.AccAddress{}, false
	}

	return bz, true
}

// GetCosmosChainByName gets the address prefix of the given cosmos chain
func (k Keeper) GetCosmosChainByName(ctx sdk.Context, chain string) (types.CosmosChain, bool) {
	key := cosmosChainPrefix.Append(utils.LowerCaseKey(chain))
	var value types.CosmosChain
	ok := k.getStore(ctx).Get(key, &value)
	if !ok {
		return types.CosmosChain{}, false
	}

	return value, true
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
