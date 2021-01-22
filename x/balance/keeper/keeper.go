package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
)

const (
	senderPrefix   = "send_"
	pendingPrefix  = "pend_"
	archivedPrefix = "arch_"

	sequenceKey = "nextID"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

// LinkAddresses links a sender address to a crosschain recipient address
func (k Keeper) LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(marshalCrossChainAddress(sender)), k.cdc.MustMarshalBinaryLengthPrefixed(recipient))
}

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
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, amount sdk.Coin) error {
	recipient, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return fmt.Errorf("no recipient linked to sender %s", sender.String())
	}
	k.setPendingTransfer(ctx, recipient, amount)

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
	if bz != nil {
		ctx.KVStore(k.storeKey).Delete([]byte(pendingPrefix + marshalCrossChainKey(transfer.Recipient.Chain, transfer.ID)))
		ctx.KVStore(k.storeKey).Set([]byte(archivedPrefix+marshalCrossChainKey(transfer.Recipient.Chain, transfer.ID)), bz)
	}
}

func (k Keeper) setPendingTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) {

	var next uint64 = 0
	bz := ctx.KVStore(k.storeKey).Get([]byte(sequenceKey))
	if bz != nil {
		next = binary.LittleEndian.Uint64(bz)
	}

	transfer := exported.CrossChainTransfer{Recipient: recipient, Amount: amount, ID: next}
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
