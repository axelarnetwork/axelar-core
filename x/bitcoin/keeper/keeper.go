package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	rawPrefix  = "raw_"
	utxoPrefix = "utxo_"
	pollPrefix = "poll_"
	addrPrefix = "addr_"
)

var (
	confHeight = []byte("confHeight")
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewBtcKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetTrackedAddress(ctx sdk.Context, address string) {
	ctx.KVStore(k.storeKey).Set([]byte(addrPrefix+address), []byte{})
}

func (k Keeper) GetTrackedAddress(ctx sdk.Context, address string) types.ExternalChainAddress {
	val := ctx.KVStore(k.storeKey).Get([]byte(addrPrefix + address))
	if val == nil {
		return ""
	}
	return address
}

func (k Keeper) SetConfirmationHeight(ctx sdk.Context, height uint64) {
	ctx.KVStore(k.storeKey).Set(confHeight, k.cdc.MustMarshalBinaryLengthPrefixed(height))
}

func (k Keeper) GetConfirmationHeight(ctx sdk.Context) uint64 {
	rawHeight := ctx.KVStore(k.storeKey).Get(confHeight)
	if rawHeight == nil {
		return types.DefaultGenesisState().ConfirmationHeight
	} else {
		var height uint64
		k.cdc.MustUnmarshalBinaryLengthPrefixed(rawHeight, &height)
		return height
	}
}

func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

func (k Keeper) GetRawTx(ctx sdk.Context, txId string) *wire.MsgTx {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + txId))
	if bz == nil {
		return nil
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx
}

func (k Keeper) SetRawTx(ctx sdk.Context, txID string, tx *wire.MsgTx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+txID), bz)
}

func (k Keeper) setUTXO(ctx sdk.Context, txId string, utxo types.UTXO) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(utxo)
	ctx.KVStore(k.storeKey).Set([]byte(utxoPrefix+txId), bz)
}

func (k Keeper) GetUTXO(ctx sdk.Context, txID string) (types.UTXO, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(utxoPrefix + txID))
	if bz == nil {
		return types.UTXO{}, false
	}
	var utxo types.UTXO
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &utxo)

	return utxo, true
}

func (k Keeper) SetUTXOForPoll(ctx sdk.Context, pollID string, utxo types.UTXO) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(utxo)
	ctx.KVStore(k.storeKey).Set([]byte(pollPrefix+pollID), bz)
}

func (k Keeper) GetUTXOForPoll(ctx sdk.Context, pollID string) (types.UTXO, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pollPrefix + pollID))
	if bz == nil {
		return types.UTXO{}, false
	}
	var utxo types.UTXO
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &utxo)

	return utxo, true
}

// ProcessUTXOPollResult stores the UTXO permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessUTXOPollResult(ctx sdk.Context, pollID string, confirmed bool) error {
	utxo, ok := k.GetUTXOForPoll(ctx, pollID)
	if !ok {
		return fmt.Errorf("poll not found")
	}
	if confirmed {
		k.setUTXO(ctx, utxo.Hash.String(), utxo)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pollPrefix + pollID))
	return nil
}
