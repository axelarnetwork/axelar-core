package keeper

import (
	"encoding/binary"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

var (
	transferPrefix    = utils.KeyFromStr("transfer")
	cosmosChainPrefix = utils.KeyFromStr("cosmos_chain")
	feeCollector      = utils.KeyFromStr("fee_collector")
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

func (k Keeper) setParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
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
	cosmosChain, ok := k.getCosmosChain(ctx, chain)
	if !ok {
		return fmt.Errorf("unknown cosmos chain %s", chain)
	}

	if _, ok := k.GetIBCPath(ctx, chain); ok {
		return fmt.Errorf("path %s already registered for cosmos chain %s", path, chain)
	}

	cosmosChain.IBCPath = path
	k.SetCosmosChain(ctx, cosmosChain)

	return nil
}

// GetIBCPath retrieves the IBC path associated to the specified chain
func (k Keeper) GetIBCPath(ctx sdk.Context, chain string) (string, bool) {
	cosmosChain, ok := k.getCosmosChain(ctx, chain)
	if !ok || cosmosChain.IBCPath == "" {
		return "", false
	}

	return cosmosChain.IBCPath, true
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
	key := transferPrefix.Append(utils.KeyFromStr(portID)).Append(utils.KeyFromStr(channelID)).Append(utils.KeyFromBz(bz))

	k.getStore(ctx).Delete(key)
}

// IsCosmosChain returns true if the given chain name is for a cosmos chain
func (k Keeper) IsCosmosChain(ctx sdk.Context, chain string) bool {
	_, ok := k.getCosmosChain(ctx, chain)
	return ok
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

// GetCosmosChains retrieves all registered cosmos chain names
func (k Keeper) GetCosmosChains(ctx sdk.Context) []string {
	cosmosChains := k.getCosmosChains(ctx)
	chains := make([]string, len(cosmosChains))

	for i, chain := range cosmosChains {
		chains[i] = chain.Name
	}

	return chains
}

func (k Keeper) getCosmosChains(ctx sdk.Context) (cosmosChains []types.CosmosChain) {
	iter := k.getStore(ctx).Iterator(cosmosChainPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var cosmosChain types.CosmosChain
		iter.UnmarshalValue(&cosmosChain)

		cosmosChains = append(cosmosChains, cosmosChain)
	}

	return cosmosChains
}

func (k Keeper) getCosmosChain(ctx sdk.Context, chain string) (cosmosChain types.CosmosChain, ok bool) {
	return cosmosChain, k.getStore(ctx).Get(cosmosChainPrefix.Append(utils.LowerCaseKey(chain)), &cosmosChain)
}

// SetCosmosChain sets the address prefix for the given cosmos chain
func (k Keeper) SetCosmosChain(ctx sdk.Context, chain types.CosmosChain) {
	// register a cosmos chain to axelarnet
	k.getStore(ctx).Set(cosmosChainPrefix.Append(utils.LowerCaseKey(chain.Name)), &chain)
}

// SetFeeCollector sets axelarnet fee collector
func (k Keeper) SetFeeCollector(ctx sdk.Context, address sdk.AccAddress) error {
	if err := sdk.VerifyAddressFormat(address); err != nil {
		return err
	}

	k.getStore(ctx).SetRaw(feeCollector, address)
	return nil
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
