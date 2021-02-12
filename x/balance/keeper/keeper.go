package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/balance/types"
)

const (
	senderPrefix   = "send_"
	infoPrefix     = "info_"
	pendingPrefix  = "pend_"
	archivedPrefix = "arch_"
	totalPrefix    = "total_"

	sequenceKey = "nextID"
)

// Keeper represents a ballance keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	params   params.Subspace
}

// NewKeeper returns a new balance keeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the balance module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)

	// By copying this data to the KV store, we avoid having to iterate across all element
	// in the parameters table when a caller needs to fetch information from it
	for _, info := range p.ChainsAssetInfo {
		k.SetChainAssetInfo(ctx, info)
	}
}

// GetParams gets the balance module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// GetChainAssetInfo retrieves the specification for a chain's assets
func (k Keeper) GetChainAssetInfo(ctx sdk.Context, chain exported.Chain) (info types.ChainAssetInfo, found bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(infoPrefix + chain.String()))
	if bz == nil {
		return
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
	found = true

	return
}

// SetChainAssetInfo sets the specification for a chain's assets
func (k Keeper) SetChainAssetInfo(ctx sdk.Context, info types.ChainAssetInfo) {
	ctx.KVStore(k.storeKey).Set([]byte(infoPrefix+info.Chain.String()), k.cdc.MustMarshalBinaryLengthPrefixed(info))
}

// LinkAddresses links a sender address to a crosschain recipient address
func (k Keeper) LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress) error {
	if _, ok := k.GetChainAssetInfo(ctx, sender.Chain); !ok {
		return fmt.Errorf("no chain asset info available for sender %s", sender.String())
	}
	if _, ok := k.GetChainAssetInfo(ctx, recipient.Chain); !ok {
		return fmt.Errorf("no chain asset info available for recipient %s", recipient.String())
	}

	ctx.KVStore(k.storeKey).Set([]byte(marshalCrossChainAddress(sender)), k.cdc.MustMarshalBinaryLengthPrefixed(recipient))
	return nil
}

// GetRecipient retrieves the cross chain recipient associated to the specified sender
func (k Keeper) GetRecipient(ctx sdk.Context, sender exported.CrossChainAddress) (exported.CrossChainAddress, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(marshalCrossChainAddress(sender)))
	if bz == nil {
		return exported.CrossChainAddress{}, false
	}

	var recp exported.CrossChainAddress
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &recp)
	return recp, true
}

// EnqueueForTransfer appoints the amount of tokens to be transfered/minted to the recipient previously linked to the specified sender
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin) error {
	infoSender, ok := k.GetChainAssetInfo(ctx, sender.Chain)
	if !ok {
		return fmt.Errorf("no chain asset info available for sender %s", sender.String())
	}
	if !infoSender.SupportsForeignAssets && infoSender.NativeAsset != asset.Denom {
		return fmt.Errorf("sender's chain %s does not support foreign assets", sender.Chain.String())
	}

	if infoSender.NativeAsset != asset.Denom && !k.getChainTotal(ctx, sender.Chain, asset.Denom).IsGTE(asset) {
		return fmt.Errorf("not enough funds available for asset '%s' in chain %s", asset.Denom, sender.Chain)
	}

	recipient, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return fmt.Errorf("no recipient linked to sender %s", sender.String())
	}

	infoRecipient, _ := k.GetChainAssetInfo(ctx, recipient.Chain)
	if !infoRecipient.SupportsForeignAssets && infoRecipient.NativeAsset != asset.Denom {
		return fmt.Errorf("recipient's chain %s does not support foreign assets", recipient.Chain.String())
	}

	if infoSender.NativeAsset != asset.Denom {
		k.subChainTotal(ctx, sender.Chain, asset)
	}
	k.setPendingTransfer(ctx, recipient, asset)
	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s to cross chain address %s in %s successfully prepared",
		asset.Amount.String(), recipient.Address, recipient.Chain.String()))

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
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + marshalCrossChainKey(transfer.Recipient.Chain, transfer.ID)))
	if bz == nil {
		return
	}

	// Archive the transfer
	ctx.KVStore(k.storeKey).Delete([]byte(pendingPrefix + marshalCrossChainKey(transfer.Recipient.Chain, transfer.ID)))
	ctx.KVStore(k.storeKey).Set([]byte(archivedPrefix+marshalCrossChainKey(transfer.Recipient.Chain, transfer.ID)), bz)

	// Update the total balance for the chain if it is a foreign asset
	var t exported.CrossChainTransfer
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &t)
	info, _ := k.GetChainAssetInfo(ctx, t.Recipient.Chain)
	if info.NativeAsset != t.Asset.Denom {
		k.addChainTotal(ctx, t.Recipient.Chain, t.Asset)
	}
}

func (k Keeper) getChainTotal(ctx sdk.Context, chain exported.Chain, denom string) sdk.Coin {
	bz := ctx.KVStore(k.storeKey).Get([]byte(totalPrefix + chain.String() + "_" + denom))
	if bz == nil {
		return sdk.NewCoin(denom, sdk.ZeroInt())
	}

	var total sdk.Coin
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &total)
	return total
}

func (k Keeper) addChainTotal(ctx sdk.Context, chain exported.Chain, amount sdk.Coin) {
	total := k.getChainTotal(ctx, chain, amount.Denom)
	total = total.Add(amount)

	ctx.KVStore(k.storeKey).Set([]byte(totalPrefix+chain.String()+"_"+amount.Denom), k.cdc.MustMarshalBinaryLengthPrefixed(total))
}

func (k Keeper) subChainTotal(ctx sdk.Context, chain exported.Chain, withdrawal sdk.Coin) {
	total := k.getChainTotal(ctx, chain, withdrawal.Denom)
	total = total.Sub(withdrawal)

	ctx.KVStore(k.storeKey).Set([]byte(totalPrefix+chain.String()+"_"+withdrawal.Denom), k.cdc.MustMarshalBinaryLengthPrefixed(total))
}

func (k Keeper) setPendingTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) {

	var next uint64 = 0
	bz := ctx.KVStore(k.storeKey).Get([]byte(sequenceKey))
	if bz != nil {
		next = binary.LittleEndian.Uint64(bz)
	}

	transfer := exported.CrossChainTransfer{Recipient: recipient, Asset: amount, ID: next}
	ctx.KVStore(k.storeKey).Set([]byte(pendingPrefix+marshalCrossChainKey(recipient.Chain, next)), k.cdc.MustMarshalBinaryLengthPrefixed(transfer))

	next++
	bz = make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, next)
	ctx.KVStore(k.storeKey).Set([]byte(sequenceKey), bz)
}

func (k Keeper) getAddresses(ctx sdk.Context, getType string, chain exported.Chain) []exported.CrossChainTransfer {
	transfers := make([]exported.CrossChainTransfer, 0)
	prefix := []byte(getType + chain.String() + "_")

	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), prefix)
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var transfer exported.CrossChainTransfer
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &transfer)
		transfers = append(transfers, transfer)
	}

	return transfers
}

func marshalCrossChainAddress(addr exported.CrossChainAddress) string {
	return senderPrefix + addr.Chain.String() + "_" + addr.Address
}

func marshalCrossChainKey(chain exported.Chain, sequence uint64) string {
	return fmt.Sprintf("%s_%d", chain.String(), sequence)
}
