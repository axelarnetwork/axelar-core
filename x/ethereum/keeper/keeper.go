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
	pollPrefix      = "poll_"
	scPrefix        = "sc_"
	txIDForSCPrefix = "sctxID_"
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

func (k Keeper) GetRawTx(ctx sdk.Context, txID string) *ethTypes.Transaction {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + txID))
	if bz == nil {
		return nil
	}
	var tx *ethTypes.Transaction
	k.cdc.MustUnmarshalJSON(bz, &tx)

	return tx
}

func (k Keeper) SetRawTx(ctx sdk.Context, txID string, tx *ethTypes.Transaction) {
	bz := k.cdc.MustMarshalJSON(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+txID), bz)
}

func (k Keeper) setTx(ctx sdk.Context, txID string, tx types.Tx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(txPrefix+txID), bz)
}

func (k Keeper) GetTx(ctx sdk.Context, txId string) (types.Tx, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(txPrefix + txId))
	if bz == nil {
		return types.Tx{}, false
	}
	var tx types.Tx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx, true
}

func (k Keeper) SetTxForPoll(ctx sdk.Context, pollID string, tx types.Tx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(pollPrefix+pollID), bz)
}

func (k Keeper) GetTxForPoll(ctx sdk.Context, pollID string) (types.Tx, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pollPrefix + pollID))
	if bz == nil {
		return types.Tx{}, false
	}
	var tx types.Tx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx, true
}

// ProcessTxPollResult stores the TX permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessTxPollResult(ctx sdk.Context, pollID string, confirmed bool) error {
	tx, ok := k.GetTxForPoll(ctx, pollID)
	if !ok {
		return fmt.Errorf("poll not found")
	}
	if confirmed {
		k.setTx(ctx, tx.Hash.String(), tx)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pollPrefix + pollID))
	return nil
}

func (k Keeper) SetSmartContract(ctx sdk.Context, scId string, bytecode []byte) {

	ctx.KVStore(k.storeKey).Set([]byte(scPrefix+scId), bytecode)

}

func (k Keeper) GetSmartContract(ctx sdk.Context, scId string) []byte {

	return ctx.KVStore(k.storeKey).Get([]byte(scPrefix + scId))

}

func (k Keeper) GettxIDForContractID(ctx sdk.Context, contractID string, networkID types.Network) (common.Hash, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(txIDForSCPrefix + contractID + string(networkID)))
	if bz == nil {
		return common.Hash{}, false
	}
	return common.BytesToHash(bz), true
}

func (k Keeper) SettxIDForContractID(ctx sdk.Context, contractID string, networkID types.Network, txID common.Hash) {
	ctx.KVStore(k.storeKey).Set([]byte(txIDForSCPrefix+contractID+string(networkID)), txID.Bytes())
}
