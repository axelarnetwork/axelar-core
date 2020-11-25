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

const (
	lockingPeriodKey = "lockingPeriod"
)

type Keeper struct {
	broadcaster   types.Broadcaster
	stakingKeeper types.Staker // needed only for `GetAllValidators`
	client        tssd.GG18Client
	keygenStreams map[string]tssd.GG18_KeygenClient
	signStreams   map[string]tssd.GG18_SignClient
	paramSpace    params.Subspace
}

func NewKeeper(client tssd.GG18Client, paramSpace params.Subspace, broadcaster types.Broadcaster, staking types.Staker) Keeper {
	return Keeper{
		broadcaster:   broadcaster,
		stakingKeeper: staking,
		client:        client,
		keygenStreams: map[string]tssd.GG18_KeygenClient{},
		signStreams:   map[string]tssd.GG18_SignClient{},
		paramSpace:    paramSpace,
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

func (k Keeper) IsKeyRefreshLocked(ctx sdk.Context, snapshotTime time.Time) bool {
	lp := k.lockingPeriod(ctx)
	return snapshotTime.Add(lp).Before(ctx.BlockTime())
}

func (k Keeper) lockingPeriod(ctx sdk.Context) (lockingPeriod time.Duration) {
	k.paramSpace.Get(ctx, []byte(lockingPeriodKey), &lockingPeriod)
	return
}
