package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/axelar/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	bridges  map[string]types.BridgeKeeper
	tss      types.TSSKeeper
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, bridges map[string]types.BridgeKeeper, tss types.TSSKeeper) Keeper {
	keeper := Keeper{
		bridges:  bridges,
		storeKey: key,
		cdc:      cdc,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) TrackAddress(ctx sdk.Context, address types.ExternalChainAddress) error {
	br, ok := k.bridges[address.Chain]
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Chain with name % is not bridged.", address.Chain)
	}

	if err := br.TrackAddress(ctx, address.Address); err != nil {
		k.Logger(ctx).Info(sdkerrors.Wrapf(err, "Bridge to %s is unable to track address", address.Chain).Error())
		return sdkerrors.Wrapf(err, "Bridge to %s is unable to track address", address.Chain)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Bridge to %s is able to track address", address.Chain))

	ctx.KVStore(k.storeKey).Set([]byte(address.Address), []byte(address.Chain))

	return nil
}

func (k Keeper) GetTrackedAddress(ctx sdk.Context, address string) types.ExternalChainAddress {
	chain := ctx.KVStore(k.storeKey).Get([]byte(address))
	if chain == nil {
		return types.ExternalChainAddress{}
	}
	return types.ExternalChainAddress{
		Chain:   string(chain),
		Address: address,
	}
}
