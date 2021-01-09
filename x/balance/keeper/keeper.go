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
	pending         = "pend_"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

func (k Keeper) LinkAddresses(ctx sdk.Context, sender exported.CrossChainAddress, recipient exported.CrossChainAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(senderPrefix+sender.String()), k.cdc.MustMarshalBinaryBare(recipient))
	senders := k.GetSenders(ctx, recipient)
	senders = append(senders, sender)
	ctx.KVStore(k.storeKey).Set([]byte(recipientPrefix+recipient.String()), k.cdc.MustMarshalBinaryBare(senders))
}

func (k Keeper) PrepareForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, amount sdk.Coin) error {
	recp, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return fmt.Errorf("no recipient linked to sender %s", sender.String())
	}
	transfers := k.GetPendingTransfers(ctx, recp)
	transfers = transfers.Add(amount)
	ctx.KVStore(k.storeKey).Set([]byte(pending+recp.String()), k.cdc.MustMarshalBinaryBare(transfers))
	return nil
}

func (k Keeper) GetRecipient(ctx sdk.Context, sender exported.CrossChainAddress) (exported.CrossChainAddress, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(senderPrefix + sender.String()))
	if bz == nil {
		return exported.CrossChainAddress{}, false
	}
	var recp exported.CrossChainAddress
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &recp)
	return recp, true
}

func (k Keeper) GetPendingTransfers(ctx sdk.Context, recipient exported.CrossChainAddress) sdk.Coins {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pending + recipient.String()))

	if bz == nil {
		return sdk.NewCoins()
	}
	var transfers sdk.Coins
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &transfers)
	return transfers
}

func (k Keeper) GetSenders(ctx sdk.Context, recipient exported.CrossChainAddress) []exported.CrossChainAddress {
	bz := ctx.KVStore(k.storeKey).Get([]byte(recipientPrefix + recipient.String()))
	if bz == nil {
		return nil
	}
	var senders []exported.CrossChainAddress
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &senders)
	return senders
}
