package voting

import (
	"github.com/axelarnetwork/axelar-core/x/voting/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, k keeper.Keeper) []abci.ValidatorUpdate {
	if req.Height%k.GetVotingInterval(ctx) == 0 {
		k.TallyVotes(ctx)
		// if voting fails the votes will be counted as discards, no point in handling that here
		err := k.BatchVote(ctx)
		if err != nil {
			k.Logger(ctx).Error(err.Error())
		}
	}
	return nil
}
