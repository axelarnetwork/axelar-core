package keeper

import (
	"context"
	"fmt"
	"time"

	tssd "github.com/axelarnetwork/tssd/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

type Keeper struct {
	broadcaster   types.Broadcaster
	staker        types.Staker
	client        tssd.GG18Client
	keygenStreams map[string]tssd.GG18_KeygenClient
	signStreams   map[string]tssd.GG18_SignClient
	paramSpace    params.Subspace
	voter         types.Voter
	storeKey      sdk.StoreKey
}

func NewKeeper(storeKey sdk.StoreKey, client tssd.GG18Client, paramSpace params.Subspace, broadcaster types.Broadcaster, staking types.Staker, voter types.Voter) Keeper {
	return Keeper{
		broadcaster:   broadcaster,
		staker:        staking,
		voter:         voter,
		client:        client,
		keygenStreams: map[string]tssd.GG18_KeygenClient{},
		signStreams:   map[string]tssd.GG18_SignClient{},
		paramSpace:    paramSpace.WithKeyTable(types.KeyTable()),
		storeKey:      storeKey,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// need to create a new context for every new protocol start
func (k Keeper) newContext() (context.Context, context.CancelFunc) {
	// TODO: make timeout a config parameter?
	return context.WithTimeout(context.Background(), 2*time.Hour)
}

// IsKeyRefreshLocked checks if the master key swap is currently locked
func (k Keeper) IsKeyRefreshLocked(ctx sdk.Context, snapshotHeight int64) bool {
	p := k.GetParams(ctx)
	return snapshotHeight+p.LockingPeriod > ctx.BlockHeight()
}

// SetParams sets the tss module's parameters
func (k Keeper) SetParams(ctx sdk.Context, set types.Params) {
	k.paramSpace.SetParamSet(ctx, &set)
}

// SetParams gets the tss module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}
