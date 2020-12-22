package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

var (
	confHeight = []byte("confHeight")
)

const (
	rawKey = "raw"
	txKey  = "utxo"
	scKey  = "utxo"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewEthKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey}
}

func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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

func (k Keeper) GetRawTx(ctx sdk.Context, txId string) *ethTypes.Transaction {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawKey + txId))
	if bz == nil {
		return nil
	}
	var tx *ethTypes.Transaction
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx
}

func (k Keeper) SetRawTx(ctx sdk.Context, txId string, tx *ethTypes.Transaction) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawKey+txId), bz)
}

func (k Keeper) SetTX(ctx sdk.Context, txId string, utxo types.TX) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(utxo)
	ctx.KVStore(k.storeKey).Set([]byte(txKey+txId), bz)
}

func (k Keeper) GetTX(ctx sdk.Context, txId string) (types.TX, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(txKey + txId))
	if bz == nil {
		return types.TX{}, false
	}
	var tx types.TX
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx, true
}

func (k Keeper) SetSmartContract(ctx sdk.Context, scId string, bytecode []byte) {

	ctx.KVStore(k.storeKey).Set([]byte(scKey+scId), bytecode)

}

func (k Keeper) GetSmartContract(ctx sdk.Context, scId string) []byte {

	return ctx.KVStore(k.storeKey).Get([]byte(txKey + scId))

}
