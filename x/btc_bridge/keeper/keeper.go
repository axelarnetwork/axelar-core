package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
)

const (
	rawKey  = "raw"
	utxoKey = "utxo"
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
	ctx.KVStore(k.storeKey).Set([]byte(address), []byte{})
}

func (k Keeper) GetTrackedAddress(ctx sdk.Context, address string) exported.ExternalChainAddress {
	val := ctx.KVStore(k.storeKey).Get([]byte(address))
	if val == nil {
		return exported.ExternalChainAddress{}
	}
	return exported.ExternalChainAddress{
		Chain:   "Bitcoin",
		Address: address,
	}
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
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawKey + txId))
	if bz == nil {
		return nil
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, tx)

	return tx
}

func (k Keeper) SetRawTx(ctx sdk.Context, txId string, tx *wire.MsgTx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawKey+txId), bz)
}

type serializableUtxo struct {
	Hash    *chainhash.Hash
	VoutIdx uint32
	Amount  btcutil.Amount
	Address string
	Chain   string
}

func (k Keeper) SetUTXO(ctx sdk.Context, txId string, utxo types.UTXO) {
	sUtxo := serializableUtxo{
		Hash:    utxo.Hash,
		VoutIdx: utxo.VoutIdx,
		Amount:  utxo.Amount,
		Address: utxo.Address.EncodeAddress(),
	}
	if utxo.Address.IsForNet(&chaincfg.MainNetParams) {
		sUtxo.Chain = chaincfg.MainNetParams.Name
	}
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(sUtxo)
	ctx.KVStore(k.storeKey).Set([]byte(utxoKey+txId), bz)
}

func (k Keeper) GetUTXO(ctx sdk.Context, txId string) (types.UTXO, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(utxoKey + txId))
	if bz == nil {
		return types.UTXO{}, false
	}
	var sUtxo serializableUtxo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &sUtxo)
	var chain *chaincfg.Params
	if sUtxo.Chain == chaincfg.MainNetParams.Name {
		chain = &chaincfg.MainNetParams
	} else {
		chain = &chaincfg.TestNet3Params
	}

	address, _ := btcutil.DecodeAddress(sUtxo.Address, chain)
	return types.UTXO{
		Hash:    sUtxo.Hash,
		VoutIdx: sUtxo.VoutIdx,
		Amount:  sUtxo.Amount,
		Address: address,
	}, true
}
