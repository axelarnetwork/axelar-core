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
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
)

var (
	cosmosChainPrefix = utils.KeyFromStr("cosmos_chain")
	feeCollector      = utils.KeyFromStr("fee_collector")

	nonceKey             = utils.KeyFromStr("nonce")
	transferPrefix       = utils.KeyFromStr("ibc_transfer")
	ibcTransferQueueName = "ibc_transfer_queue"
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

// RegisterIBCPath registers an IBC path for a cosmos chain
func (k Keeper) RegisterIBCPath(ctx sdk.Context, chain nexus.ChainName, path string) error {
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
func (k Keeper) GetIBCPath(ctx sdk.Context, chain nexus.ChainName) (string, bool) {
	cosmosChain, ok := k.getCosmosChain(ctx, chain)
	if !ok || cosmosChain.IBCPath == "" {
		return "", false
	}

	return cosmosChain.IBCPath, true
}

// IsCosmosChain returns true if the given chain name is for a cosmos chain
func (k Keeper) IsCosmosChain(ctx sdk.Context, chain nexus.ChainName) bool {
	_, ok := k.getCosmosChain(ctx, chain)
	return ok
}

// GetCosmosChainByName gets the address prefix of the given cosmos chain
func (k Keeper) GetCosmosChainByName(ctx sdk.Context, chain nexus.ChainName) (types.CosmosChain, bool) {
	key := cosmosChainPrefix.Append(utils.LowerCaseKey(chain.String()))
	var value types.CosmosChain
	ok := k.getStore(ctx).Get(key, &value)
	if !ok {
		return types.CosmosChain{}, false
	}

	return value, true
}

// GetCosmosChains retrieves all registered cosmos chain names
func (k Keeper) GetCosmosChains(ctx sdk.Context) []nexus.ChainName {
	return slices.Map(k.getCosmosChains(ctx), func(c types.CosmosChain) nexus.ChainName { return c.Name })
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

func (k Keeper) getCosmosChain(ctx sdk.Context, chain nexus.ChainName) (cosmosChain types.CosmosChain, ok bool) {
	return cosmosChain, k.getStore(ctx).Get(cosmosChainPrefix.Append(utils.LowerCaseKey(chain.String())), &cosmosChain)
}

// SetCosmosChain sets the address prefix for the given cosmos chain
func (k Keeper) SetCosmosChain(ctx sdk.Context, chain types.CosmosChain) {
	// register a cosmos chain to axelarnet
	k.getStore(ctx).Set(cosmosChainPrefix.Append(utils.LowerCaseKey(chain.Name.String())), &chain)
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

// GetIBCTransferQueue returns the queue of IBC transfers
func (k Keeper) GetIBCTransferQueue(ctx sdk.Context) utils.KVQueue {
	return utils.NewGeneralKVQueue(
		ibcTransferQueueName,
		k.getStore(ctx),
		k.Logger(ctx),
		func(value codec.ProtoMarshaler) utils.Key {
			transfer := value.(*types.IBCTransfer)
			return utils.KeyFromBz(transfer.ID.Bytes())
		},
	)
}

// EnqueueTransfer stores the pending ibc transfer in the queue
func (k Keeper) EnqueueTransfer(ctx sdk.Context, transfer types.IBCTransfer) error {
	id := k.getNonce(ctx)
	transfer.SetID(id)
	k.setNonce(ctx, id+1)

	key := transferPrefix.AppendStr(transfer.ID.String())
	if k.getStore(ctx).Has(key) {
		return fmt.Errorf("transfer %s already exists", transfer.ID.String())
	}

	k.GetIBCTransferQueue(ctx).Enqueue(key, &transfer)
	return nil
}

func (k Keeper) getNonce(ctx sdk.Context) uint64 {
	if bz := k.getStore(ctx).GetRaw(nonceKey); bz != nil {
		return binary.LittleEndian.Uint64(bz)
	}

	return 0
}

func (k Keeper) setNonce(ctx sdk.Context, nonce uint64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, nonce)

	k.getStore(ctx).SetRaw(nonceKey, bz)
}

// validateIBCTransferQueueState checks if the keys of the given map have the correct format to be imported as ibc transfer queue state.
func (k Keeper) validateIBCTransferQueueState(state utils.QueueState, queueName ...string) error {
	if err := state.ValidateBasic(queueName...); err != nil {
		return err
	}

	for _, item := range state.Items {
		var transfer types.IBCTransfer
		if err := k.cdc.UnmarshalLengthPrefixed(item.Value, &transfer); err != nil {
			return err
		}

		if err := transfer.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
