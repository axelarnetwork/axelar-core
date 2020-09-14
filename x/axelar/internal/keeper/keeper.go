package keeper

import (
	"fmt"
	"github.com/axelarnetwork/axelar-net/x/axelar/internal/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Bridge interface {
	TrackAddress(address []byte) error
}

type Keeper struct {
	bridges  map[string]Bridge
	storeKey sdk.StoreKey
	cdc      *codec.Codec
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, bridges map[string]Bridge) Keeper {
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
	bridge, ok := k.bridges[address.Chain]
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Chain with name % is not bridged.", address.Chain)
	}

	if err := bridge.TrackAddress(address.Address); err != nil {
		return sdkerrors.Wrapf(err, "Bridge to %s is unable to track address", address.Chain)
	}

	ctx.KVStore(k.storeKey).Set(address.Address, []byte(address.Chain))

	return nil
}

func (k Keeper) GetTrackedAddress(ctx sdk.Context, address []byte) types.ExternalChainAddress {
	chain := ctx.KVStore(k.storeKey).Get(address)
	if chain == nil {
		return types.ExternalChainAddress{}
	}
	return types.ExternalChainAddress{
		Chain:   string(chain),
		Address: address,
	}
}
