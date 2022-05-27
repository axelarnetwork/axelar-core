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
	hasPollExpired := func(value codec.ProtoMarshaler) (bool, bool) {
		return ctx.BlockHeight() >= value.(*exported.PollMetadata).ExpiresAt, true
	}

	var pollMeta exported.PollMetadata
	for pollQueue.Dequeue(&pollMeta, hasPollExpired) {
		poll := k.GetPoll(ctx, pollMeta.Key)

		voteHandler := k.GetVoteRouter().GetHandler(poll.GetKey().Module)
		if voteHandler == nil {
			return fmt.Errorf("unknown module for vote %s", poll.GetKey().Module)
		}

		switch {
		case poll.Is(exported.Pending):
			poll.SetExpired()
			poll.AllowOverride()

			if err := voteHandler.HandleExpiredPoll(ctx, poll); err != nil {
				return err
			}

			k.Logger(ctx).Debug("marked poll as expired and can be re-started now",
				"poll", poll.GetKey().String(),
			)
		case poll.Is(exported.Failed):
			poll.AllowOverride()

			k.Logger(ctx).Debug("poll failed and can be re-started now",
				"poll", poll.GetKey().String(),
			)
		case poll.Is(exported.Completed):
			if err := voteHandler.HandleCompletedPoll(ctx, poll); err != nil {
				return err
			}

			if voteHandler.IsFalsyResult(poll.GetResult()) {
				poll.AllowOverride()
				k.Logger(ctx).Debug("poll completed with falsy result and can be re-started now",
					"poll", poll.GetKey().String(),
				)
			} else {
				k.Logger(ctx).Debug("poll completed with final result",
					"poll", poll.GetKey().String(),
				)
			}
		default:
			return fmt.Errorf("cannot handle poll %s due to invalid state", poll.GetKey())
		}
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
