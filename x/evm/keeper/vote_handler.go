package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/slices"
)

var _ vote.VoteHandler = &voteHandler{}

type voteHandler struct {
	cdc      codec.Codec
	keeper   types.BaseKeeper
	nexus    types.Nexus
	rewarder types.Rewarder
}

// NewVoteHandler returns the handler for processing vote delivered by the vote module
func NewVoteHandler(cdc codec.Codec, keeper types.BaseKeeper, nexus types.Nexus, rewarder types.Rewarder) vote.VoteHandler {
	return voteHandler{
		cdc:      cdc,
		keeper:   keeper,
		nexus:    nexus,
		rewarder: rewarder,
	}
}

func (v voteHandler) IsFalsyResult(result codec.ProtoMarshaler) bool {
	voteEvents, err := types.UnpackEvents(v.cdc, result.(*vote.Vote).Result)

	return err != nil || len(voteEvents.Events) == 0
}

func (v voteHandler) HandleExpiredPoll(ctx sdk.Context, poll vote.Poll) error {
	rewardPoolName, ok := poll.GetRewardPoolName()
	if !ok {
		return fmt.Errorf("reward pool not set for poll %s", poll.GetKey().String())
	}

	// TODO: MarkChainMaintainerMissingVote for those who didn't vote in time. Need
	// to be able to get chain of expired polls in order to do it.
	rewardPool := v.rewarder.GetPool(ctx, rewardPoolName)
	// Penalize voters who failed to vote
	for _, voter := range poll.GetVoters() {
		if !poll.HasVoted(voter.Validator) {
			rewardPool.ClearRewards(voter.Validator)
			v.keeper.Logger(ctx).Debug("penalized voter due to timeout",
				"voter", voter.Validator.String(),
				"poll", poll.GetKey().String())
		}
	}

	return nil
}

func (v voteHandler) HandleCompletedPoll(ctx sdk.Context, poll vote.Poll) error {
	voteEvents, err := types.UnpackEvents(v.cdc, poll.GetResult().(*vote.Vote).Result)
	if err != nil {
		return err
	}

	chain, ok := v.nexus.GetChain(ctx, voteEvents.Chain)
	if !ok {
		return fmt.Errorf("%s is not a registered chain", voteEvents.Chain)
	}

	rewardPoolName, ok := poll.GetRewardPoolName()
	if !ok {
		return fmt.Errorf("reward pool not set for poll %s", poll.GetKey().String())
	}

	rewardPool := v.rewarder.GetPool(ctx, rewardPoolName)

	for _, voter := range poll.GetVoters() {
		hasVoted := poll.HasVoted(voter.Validator)
		hasVotedLate := poll.HasVotedLate(voter.Validator)
		hasVotedIncorrectly := hasVoted && !poll.HasVotedCorrectly(voter.Validator)

		v.nexus.MarkChainMaintainerMissingVote(ctx, chain, voter.Validator, !hasVoted)
		v.nexus.MarkChainMaintainerIncorrectVote(ctx, chain, voter.Validator, hasVotedIncorrectly)

		v.keeper.Logger(ctx).Debug("marked voter misbehave",
			"voter", voter.Validator.String(),
			"missing_vote", !hasVoted,
			"incorrect_vote", hasVotedIncorrectly,
		)

		switch {
		case hasVotedIncorrectly:
			rewardPool.ClearRewards(voter.Validator)
			v.keeper.Logger(ctx).Debug("penalized voter due to incorrect vote",
				"voter", voter.Validator.String(),
				"poll", poll.GetKey().String())
		case hasVoted && !hasVotedLate:
			if err := rewardPool.ReleaseRewards(voter.Validator); err != nil {
				return err
			}
			v.keeper.Logger(ctx).Debug("released rewards for voter",
				"voter", voter.Validator.String(),
				"poll", poll.GetKey().String())
		default:
			v.keeper.Logger(ctx).Debug("held rewards for voter due to missing or late vote",
				"voter", voter.Validator.String(),
				"poll", poll.GetKey().String())
		}
	}

	return nil
}

func (v voteHandler) HandleResult(ctx sdk.Context, result codec.ProtoMarshaler) error {
	voteEvents, err := types.UnpackEvents(v.cdc, result.(*vote.Vote).Result)
	if err != nil {
		return err
	}

	if v.IsFalsyResult(result) {
		return nil
	}

	if slices.Any(voteEvents.Events, func(event types.Event) bool { return event.Chain != voteEvents.Chain }) {
		return fmt.Errorf("events are not from the same source chain")
	}

	chain, ok := v.nexus.GetChain(ctx, voteEvents.Chain)
	if !ok {
		return fmt.Errorf("%s is not a registered chain", voteEvents.Chain)
	}

	if !v.keeper.HasChain(ctx, voteEvents.Chain) {
		return fmt.Errorf("%s is not an evm chain", voteEvents.Chain)
	}

	chainK := v.keeper.ForChain(chain.Name)
	cacheCtx, writeCache := ctx.CacheContext()

	err = handleEvents(cacheCtx, chainK, voteEvents.Events, chain)
	if err != nil {
		// set events to failed, we will deal with later
		for _, e := range voteEvents.Events {
			chainK.SetFailedEvent(ctx, e)
		}
		return err
	}

	writeCache()
	ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())

	return nil
}

func handleEvents(ctx sdk.Context, ck types.ChainKeeper, events []types.Event, chain nexus.Chain) error {
	for _, event := range events {
		// validate event
		if err := event.ValidateBasic(); err != nil {
			return fmt.Errorf("event %s: %s", event.GetID(), err.Error())
		}

		// check if event confirmed before
		eventID := event.GetID()
		if _, ok := ck.GetEvent(ctx, eventID); ok {
			return fmt.Errorf("event %s is already confirmed", eventID)
		}
		ck.SetConfirmedEvent(ctx, event)
		ck.Logger(ctx).Info(fmt.Sprintf("confirmed %s event %s in transaction %s", chain.Name, eventID, event.TxId.Hex()))

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(types.EventTypeEventConfirmation,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyChain, event.Chain),
				sdk.NewAttribute(types.AttributeKeyTxID, event.TxId.Hex()),
				sdk.NewAttribute(types.AttributeKeyEventID, string(event.GetID())),
				sdk.NewAttribute(types.AttributeKeyEventType, event.GetEventType()),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)),
		)
	}

	return nil
}
