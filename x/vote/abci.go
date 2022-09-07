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

	endBlockerLimit := k.GetParams(ctx).EndBlockerLimit
	handledPolls := int64(0)
	var pollMetadata exported.PollMetadata
	for handledPolls < endBlockerLimit && pollQueue.DequeueIf(&pollMetadata, hasPollExpired) {
		handledPolls++

		pollID := pollMetadata.ID
		poll, ok := k.GetPoll(ctx, pollID)
		if !ok {
			panic(fmt.Errorf("poll %s not found", pollID))
		}

		logger := k.Logger(ctx).With("poll", pollID.String())

		voteHandler := k.GetVoteRouter().GetHandler(poll.GetModule())
		switch poll.GetState() {
		case exported.Pending:
			logger.Debug("poll expired")
			if err := voteHandler.HandleExpiredPoll(ctx, poll); err != nil {
				return err
			}

		case exported.Failed:
			logger.Debug("poll failed")
			if err := voteHandler.HandleFailedPoll(ctx, poll); err != nil {
				return err
			}

		case exported.Completed:
			if voteHandler.IsFalsyResult(poll.GetResult()) {
				logger.Debug("poll completed with falsy result")
			} else {
				logger.Debug("poll completed with final result")
			}
			if err := voteHandler.HandleCompletedPoll(ctx, poll); err != nil {
				return err
			}
		default:
			panic(fmt.Errorf("unexpected poll state %s", poll.GetState().String()))
		}

		k.DeletePoll(ctx, pollID)
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
