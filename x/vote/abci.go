package vote

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

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
		logger := k.Logger(ctx).With("poll", pollID.String())

		poll, ok := k.GetPoll(ctx, pollID)
		if !ok {
			logger.Error("poll not found")
			continue
		}

		handled := utils.RunCached(ctx, k, func(cachedCtx sdk.Context) (bool, error) {
			voteHandler := k.GetVoteRouter().GetHandler(poll.GetModule())

			switch poll.GetState() {
			case exported.Pending:
				logger.Debug("poll expired")
				if err := voteHandler.HandleExpiredPoll(cachedCtx, poll); err != nil {
					return false, err
				}

			case exported.Failed:
				logger.Debug("poll failed")
				if err := voteHandler.HandleFailedPoll(cachedCtx, poll); err != nil {
					return false, err
				}

			case exported.Completed:
				if voteHandler.IsFalsyResult(poll.GetResult()) {
					logger.Debug("poll completed with falsy result")
				} else {
					logger.Debug("poll completed with final result")
				}
				if err := voteHandler.HandleCompletedPoll(cachedCtx, poll); err != nil {
					return false, err
				}

			default:
				return false, fmt.Errorf("unexpected poll state %s", poll.GetState().String())
			}

			return true, nil
		})

		if !handled {
			logger.Error("failed to handle poll at expiry, dropping poll")
		}

		k.DeletePoll(ctx, pollID)
	}

	return nil
}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k types.Voter) ([]abci.ValidatorUpdate, error) {
	return nil, handlePollsAtExpiry(ctx, k)
}
