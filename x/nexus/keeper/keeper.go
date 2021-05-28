package keeper

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

const (
	senderPrefix     = "send_"
	chainPrefix      = "chain_"
	pendingPrefix    = "pend_"
	archivedPrefix   = "arch_"
	totalPrefix      = "total_"
	registeredPrefix = "registered_"

	sequenceKey = "nextID"
	registered  = 0x01
)

// Keeper represents a ballance keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler
	params   params.Subspace
}

// NewKeeper returns a new nexus keeper
func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the nexus module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)

	for _, chain := range p.Chains {
		// By copying this data to the KV store, we avoid having to iterate across all element
		// in the parameters table when a caller needs to fetch information from it
		k.SetChain(ctx, chain)

		// Native assets can be registered at start up
		k.RegisterAsset(ctx, chain.Name, chain.NativeAsset)
	}
}

// GetParams gets the nexus module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// RegisterAsset indicates that the specified asset is supported by the given chain
func (k Keeper) RegisterAsset(ctx sdk.Context, chainName, denom string) {
	k.getStore(ctx).SetRaw(utils.LowerCaseKey(denom).WithPrefix(chainName).WithPrefix(registeredPrefix), []byte{registered})
}

// IsAssetRegistered returns true if the specified asset is supported by the given chain
func (k Keeper) IsAssetRegistered(ctx sdk.Context, chainName, denom string) bool {
	return k.getStore(ctx).GetRaw(utils.LowerCaseKey(denom).WithPrefix(chainName).WithPrefix(registeredPrefix)) != nil
}

// GetChains retrieves the specification for all supported blockchains
func (k Keeper) GetChains(ctx sdk.Context) []exported.Chain {
	var results []exported.Chain

	for iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(chainPrefix)); iter.Valid(); iter.Next() {
		var chain exported.Chain
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &chain)
		results = append(results, chain)
	}

	return results
}

// GetChain retrieves the specification for a supported blockchain
func (k Keeper) GetChain(ctx sdk.Context, chainName string) (exported.Chain, bool) {
	var chain exported.Chain
	ok := k.getStore(ctx).Get(utils.LowerCaseKey(chainName).WithPrefix(chainPrefix), &chain)

	return chain, ok
}

// SetChain sets the specification for a supported chain
func (k Keeper) SetChain(ctx sdk.Context, chain exported.Chain) {
	k.getStore(ctx).Set(utils.LowerCaseKey(chain.Name).WithPrefix(chainPrefix), &chain)
}

// LinkAddresses links a sender address to a cross-chain recipient address
func (k Keeper) LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress) {
	k.getStore(ctx).Set(utils.ToLowerCaseKey(&sender).WithPrefix(senderPrefix), &recipient)
}

// GetRecipient retrieves the cross chain recipient associated to the specified sender
func (k Keeper) GetRecipient(ctx sdk.Context, sender exported.CrossChainAddress) (exported.CrossChainAddress, bool) {
	var recp exported.CrossChainAddress
	ok := k.getStore(ctx).Get(utils.ToLowerCaseKey(&sender).WithPrefix(senderPrefix), &recp)
	return recp, ok
}

// EnqueueForTransfer appoints the amount of tokens to be transfered/minted to the recipient previously linked to the specified sender
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin) error {
	if !sender.Chain.SupportsForeignAssets && sender.Chain.NativeAsset != asset.Denom {
		return fmt.Errorf("sender's chain %s does not support foreign assets", sender.Chain.Name)
	}

	if sender.Chain.NativeAsset != asset.Denom && k.getChainTotal(ctx, sender.Chain, asset.Denom).IsLT(asset) {
		return fmt.Errorf("not enough funds available for asset '%s' in chain %s", asset.Denom, sender.Chain.Name)
	}

	recipient, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return fmt.Errorf("no recipient linked to sender %s", sender.String())
	}

	if !recipient.Chain.SupportsForeignAssets && recipient.Chain.NativeAsset != asset.Denom {
		return fmt.Errorf("recipient's chain %s does not support foreign assets", recipient.Chain.Name)
	}

	if sender.Chain.NativeAsset != asset.Denom {
		k.subtractFromChainTotal(ctx, sender.Chain, asset)
	}
	k.setPendingTransfer(ctx, recipient, asset)
	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s to cross chain address %s in %s successfully prepared",
		asset.Amount.String(), recipient.Address, recipient.Chain.Name))

	return nil
}

// GetPendingTransfersForChain returns the current set of pending transfers for a given chain
func (k Keeper) GetPendingTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer {
	return k.getAddresses(ctx, pendingPrefix, chain)
}

// GetArchivedTransfersForChain returns the history of concluded transactions to a given chain
func (k Keeper) GetArchivedTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer {
	return k.getAddresses(ctx, archivedPrefix, chain)
}

// ArchivePendingTransfer marks the transfer for the given recipient as concluded and archived
func (k Keeper) ArchivePendingTransfer(ctx sdk.Context, transfer exported.CrossChainTransfer) {
	store := k.getStore(ctx)
	key := utils.LowerCaseKey(strconv.FormatUint(transfer.ID, 10)).WithPrefix(transfer.Recipient.Chain.Name)
	bz := store.GetRaw(key.WithPrefix(pendingPrefix))
	if bz == nil {
		return
	}

	// Archive the transfer
	store.Delete(key.WithPrefix(pendingPrefix))
	store.SetRaw(key.WithPrefix(archivedPrefix), bz)

	// Update the total nexus for the chain if it is a foreign asset
	var t exported.CrossChainTransfer
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &t)
	info, _ := k.GetChain(ctx, t.Recipient.Chain.Name)
	if info.NativeAsset != t.Asset.Denom {
		k.addToChainTotal(ctx, t.Recipient.Chain, t.Asset)
	}
}

func (k Keeper) getChainTotal(ctx sdk.Context, chain exported.Chain, denom string) sdk.Coin {
	var total sdk.Coin
	ok := k.getStore(ctx).Get(utils.LowerCaseKey(denom).WithPrefix(chain.Name).WithPrefix(totalPrefix), &total)
	if !ok {
		return sdk.NewCoin(denom, sdk.ZeroInt())
	}

	return total
}

func (k Keeper) addToChainTotal(ctx sdk.Context, chain exported.Chain, amount sdk.Coin) {
	total := k.getChainTotal(ctx, chain, amount.Denom)
	total = total.Add(amount)

	k.getStore(ctx).Set(utils.LowerCaseKey(amount.Denom).WithPrefix(chain.Name).WithPrefix(totalPrefix), &total)
}

func (k Keeper) subtractFromChainTotal(ctx sdk.Context, chain exported.Chain, withdrawal sdk.Coin) {
	total := k.getChainTotal(ctx, chain, withdrawal.Denom)
	total = total.Sub(withdrawal)

	k.getStore(ctx).Set(utils.LowerCaseKey(withdrawal.Denom).WithPrefix(chain.Name).WithPrefix(totalPrefix), &total)
}

func (k Keeper) setPendingTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) {
	var next uint64
	store := k.getStore(ctx)
	bz := store.GetRaw(utils.LowerCaseKey(sequenceKey))
	if bz != nil {
		next = binary.LittleEndian.Uint64(bz)
	}

	transfer := exported.CrossChainTransfer{Recipient: recipient, Asset: amount, ID: next}
	store.Set(utils.LowerCaseKey(strconv.FormatUint(next, 10)).WithPrefix(recipient.Chain.Name).WithPrefix(pendingPrefix), &transfer)

	next++
	bz = make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, next)
	store.SetRaw(utils.LowerCaseKey(sequenceKey), bz)
}

func (k Keeper) getAddresses(ctx sdk.Context, getType string, chain exported.Chain) []exported.CrossChainTransfer {
	transfers := make([]exported.CrossChainTransfer, 0)
	prefix := []byte(getType + chain.Name + "_")

	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), prefix)
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var transfer exported.CrossChainTransfer
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &transfer)
		transfers = append(transfers, transfer)
	}

	return transfers
}

func (k Keeper) getStore(ctx sdk.Context) utils.NormalizedKVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
