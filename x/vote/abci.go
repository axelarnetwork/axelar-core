package vote

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

func handlePollsAtExpiry(ctx sdk.Context, k types.Voter) error {
	pollQueue := k.GetPollQueue(ctx)
	hasPollExpired := func(value codec.ProtoMarshaler) bool {
		return ctx.BlockHeight() >= value.(*exported.PollMetadata).ExpiresAt
	}

	var pollMeta exported.PollMetadata
	for pollQueue.DequeueIf(&pollMeta, hasPollExpired) {
		poll := k.GetPoll(ctx, pollMeta.ID)

		pollModuleMetadata := poll.GetModuleMetadata()
		voteHandler := k.GetVoteRouter().GetHandler(pollModuleMetadata.Module)
		if voteHandler == nil {
			return fmt.Errorf("unknown module for vote %s", pollModuleMetadata.Module)
		}

		switch {
		case poll.Is(exported.Pending):
			if err := voteHandler.HandleExpiredPoll(ctx, poll); err != nil {
				return err
			}

			k.Logger(ctx).Debug("poll expired",
				"poll", poll.GetID().String(),
			)
		case poll.Is(exported.Failed):
			k.Logger(ctx).Debug("poll failed",
				"poll", poll.GetID().String(),
			)
		case poll.Is(exported.Completed):
			if err := voteHandler.HandleCompletedPoll(ctx, poll); err != nil {
				return err
			}

			if voteHandler.IsFalsyResult(poll.GetResult()) {
				k.Logger(ctx).Debug("poll completed with falsy result",
					"poll", poll.GetID().String(),
				)
			} else {
				k.Logger(ctx).Debug("poll completed with final result",
					"poll", poll.GetID().String(),
				)
			}
		default:
			return fmt.Errorf("cannot handle poll %s due to invalid state", poll.GetID().String())
		}

		poll.Delete()
	}

	return nil
}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k types.Voter) ([]abci.ValidatorUpdate, error) {
	if err := handlePollsAtExpiry(ctx, k); err != nil {
		return nil, err
	}

	return nil, nil
}
