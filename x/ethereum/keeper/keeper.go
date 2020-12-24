package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"

	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

var (
	confHeight = []byte("confHeight")
)

const (
	rawPrefix       = "raw_"
	txPrefix        = "tx_"
	scPrefix        = "sc_"
	txIDForSCPrefix = "scTxID_"
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
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + txId))
	if bz == nil {
		return nil
	}
	var tx *ethTypes.Transaction
	k.cdc.MustUnmarshalJSON(bz, &tx)

	return tx
}

func (k Keeper) SetRawTx(ctx sdk.Context, txId string, tx *ethTypes.Transaction) {
	bz := k.cdc.MustMarshalJSON(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+txId), bz)
}

func (k Keeper) SetTX(ctx sdk.Context, txId string, utxo types.Tx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(utxo)
	ctx.KVStore(k.storeKey).Set([]byte(txPrefix+txId), bz)
}

func (k Keeper) GetTX(ctx sdk.Context, txId string) (types.Tx, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(txPrefix + txId))
	if bz == nil {
		return types.Tx{}, false
	}
	var tx types.Tx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx, true
}

func (k Keeper) SetSmartContract(ctx sdk.Context, scId string, bytecode []byte) {

	ctx.KVStore(k.storeKey).Set([]byte(scPrefix+scId), bytecode)

}

func (k Keeper) GetSmartContract(ctx sdk.Context, scId string) []byte {

	return ctx.KVStore(k.storeKey).Get([]byte(txPrefix + scId))

}

func (k Keeper) GetTxIDForContractID(ctx sdk.Context, contractID string, networkID types.Network) (common.Hash, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(txIDForSCPrefix + contractID + string(networkID)))
	if bz == nil {
		return common.Hash{}, false
	}
	return common.BytesToHash(bz), true
}

func (k Keeper) SetTxIDForContractID(ctx sdk.Context, contractID string, networkID types.Network, txID common.Hash) {
	ctx.KVStore(k.storeKey).Set([]byte(txIDForSCPrefix+contractID+string(networkID)), txID.Bytes())
}
