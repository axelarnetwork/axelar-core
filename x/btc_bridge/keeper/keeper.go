package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
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

func (k Keeper) SetConfirmationHeight(ctx sdk.Context, height int64) {
	ctx.KVStore(k.storeKey).Set(confHeight, k.cdc.MustMarshalBinaryLengthPrefixed(height))
}

func (k Keeper) GetConfirmationHeight(ctx sdk.Context) int64 {
	rawHeight := ctx.KVStore(k.storeKey).Get(confHeight)
	if rawHeight == nil {
		return types.DefaultGenesisState().ConfirmationHeight
	} else {
		var height int64
		k.cdc.MustUnmarshalBinaryLengthPrefixed(rawHeight, &height)
		return height
	}
}

func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}
