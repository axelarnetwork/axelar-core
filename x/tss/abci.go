package tss

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, keeper keeper.Keeper, voter types.Voter, snapshotter types.Snapshotter) []abci.ValidatorUpdate {
	keygenReqs := keeper.GetAllKeygenRequestsAtCurrentHeight(ctx)
	keeper.Logger(ctx).Info(fmt.Sprintf("processing %d keygens at height %d", len(keygenReqs), ctx.BlockHeight()))

	for _, request := range keygenReqs {
		var counter int64 = 0
		snap, found := snapshotter.GetLatestSnapshot(ctx)
		if found {
			counter = snap.Counter + 1
		}

		keeper.Logger(ctx).Info(fmt.Sprintf("linking available operations to snapshot #%d", counter))
		keeper.LinkAvailableOperatorsToCounter(ctx, request.NewKeyID, exported.AckKeygen, counter)
		keeper.DeleteAtCurrentHeight(ctx, request.NewKeyID, exported.AckKeygen)

		err := types.StartKeygen(ctx, keeper, voter, snapshotter, &request)
		if err != nil {
			keeper.Logger(ctx).Error(fmt.Sprintf("error starting keygen: %s", err.Error()))
		}
	}

	return nil
}
