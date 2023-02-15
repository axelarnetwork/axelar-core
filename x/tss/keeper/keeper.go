package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var (
	lastHeartbeatAtPrefix = key.RegisterStaticKey(types.ModuleName, 1)
)

// Keeper allows access to the broadcast state
type Keeper struct {
	params   params.Subspace
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
}

// NewKeeper constructs a tss keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{
		cdc:      cdc,
		params:   paramSpace.WithKeyTable(types.KeyTable()),
		storeKey: storeKey,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetParams sets the tss module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
}

// GetParams gets the tss module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

// GetLastHeartbeatAt returns the block height at where the last heartbeat from the given participant was received
func (k Keeper) GetLastHeartbeatAt(ctx sdk.Context, participant sdk.ValAddress) int64 {
	var value gogoprototypes.Int64Value
	if k.getStore(ctx).GetNew(lastHeartbeatAtPrefix.Append(key.FromBz(participant.Bytes())), &value) {
		return value.Value
	}

	return 0
}

// SetLastHeartbeatAt sets the block height at where the last heartbeat from the given participant was received
func (k Keeper) SetLastHeartbeatAt(ctx sdk.Context, participant sdk.ValAddress) error {
	return k.getStore(ctx).SetNewValidated(lastHeartbeatAtPrefix.Append(key.FromBz(participant.Bytes())), utils.NoValidation(&gogoprototypes.Int64Value{Value: ctx.BlockHeight()}))
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
