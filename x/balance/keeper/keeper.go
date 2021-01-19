package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/balance/exported"
)

const (
	senderPrefix    = "send_"
	recipientPrefix = "recp_"
	pendingPrefix   = "pend_"
	archivedPrefix  = "arch_"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

func (k Keeper) LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(senderPrefix+marshalCrossChainAddress(sender)), k.cdc.MustMarshalBinaryBare(recipient))
}

func (k Keeper) PrepareForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, amount sdk.Coin) error {
	recp, ok := k.getRecipient(ctx, sender)
	if !ok {
		return fmt.Errorf("no recipient linked to sender %s", sender.String())
	}

	transfers := k.getPendingTransfers(ctx, recp)
	transfers = transfers.Add(amount)
	k.setPendingTransfers(ctx, recp, transfers)
	return nil
}

func (k Keeper) GetPendingTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer {
	return k.getAddresses(ctx, pendingPrefix, chain)
}

func (k Keeper) GetArchivedTransfersForChain(ctx sdk.Context, chain exported.Chain) []exported.CrossChainTransfer {
	return k.getAddresses(ctx, archivedPrefix, chain)
}

func (k Keeper) ArchivePendingTransfers(ctx sdk.Context, recipient exported.CrossChainAddress) {
	transfers := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + marshalCrossChainAddress(recipient)))
	ctx.KVStore(k.storeKey).Delete([]byte(pendingPrefix + marshalCrossChainAddress(recipient)))
	ctx.KVStore(k.storeKey).Set([]byte(archivedPrefix+marshalCrossChainAddress(recipient)), transfers)
}

func (k Keeper) setPendingTransfers(ctx sdk.Context, recipient exported.CrossChainAddress, transfers sdk.Coins) {
	ctx.KVStore(k.storeKey).Set([]byte(pendingPrefix+marshalCrossChainAddress(recipient)), k.cdc.MustMarshalBinaryBare(transfers))
}

func (k Keeper) getRecipient(ctx sdk.Context, sender exported.CrossChainAddress) (exported.CrossChainAddress, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(senderPrefix + marshalCrossChainAddress(sender)))
	if bz == nil {
		return exported.CrossChainAddress{}, false
	}

	var recp exported.CrossChainAddress
	k.cdc.MustUnmarshalBinaryBare(bz, &recp)
	return recp, true
}

func (k Keeper) getPendingTransfers(ctx sdk.Context, recipient exported.CrossChainAddress) sdk.Coins {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + marshalCrossChainAddress(recipient)))
	if bz == nil {
		return sdk.NewCoins()
	}

	var transfers sdk.Coins
	k.cdc.MustUnmarshalBinaryBare(bz, &transfers)
	return transfers
}

func (k Keeper) getAddresses(ctx sdk.Context, getType string, chain exported.Chain) []exported.CrossChainTransfer {
	transfers := make([]exported.CrossChainTransfer, 0)
	prefix := []byte(getType + chain.String() + "_")

	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), prefix)
	for ; iter.Valid(); iter.Next() {

		key := iter.Key()
		bytes := key[len(prefix)-1:]
		recipient := exported.CrossChainAddress{Address: string(bytes), Chain: chain}

		bz := iter.Value()
		var amount sdk.Coins
		k.cdc.MustUnmarshalBinaryBare(bz, &amount)

		transfers = append(transfers, exported.CrossChainTransfer{Recipient: recipient, Amount: amount})

	}

	return transfers
}

func marshalCrossChainAddress(addr exported.CrossChainAddress) string {
	return addr.Chain.String() + "_" + addr.Address
}
