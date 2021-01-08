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
	rawPrefix      = "raw_"
	outPointPrefix = "out_"
	pollPrefix     = "poll_"
	addrPrefix     = "addr_"
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

func (k Keeper) GetTrackedAddress(ctx sdk.Context, address string) string {
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

func (k Keeper) GetRawTx(ctx sdk.Context, txID string) *wire.MsgTx {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + txID))
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

func (k Keeper) setOutpointInfo(ctx sdk.Context, txID string, info types.OutPointInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(outPointPrefix+txID), bz)
}

func (k Keeper) GetVerifiedOutPoint(ctx sdk.Context, txID string) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(outPointPrefix + txID))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var out types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &out)

	return out, true
}

func (k Keeper) SetUnverifiedOutpoint(ctx sdk.Context, txID string, info types.OutPointInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(pollPrefix+txID), bz)
}

func (k Keeper) GetUnverifiedOutPoint(ctx sdk.Context, txID string) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pollPrefix + txID))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var info types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)

	return info, true
}

// ProcessVerificationResult stores the OutPointInfo permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessVerificationResult(ctx sdk.Context, txID string, verified bool) error {
	info, ok := k.GetUnverifiedOutPoint(ctx, txID)
	if !ok {
		return fmt.Errorf("poll not found")
	}
	if verified {
		k.setOutpointInfo(ctx, info.OutPoint.Hash.String(), info)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pollPrefix + txID))
	return nil
}
