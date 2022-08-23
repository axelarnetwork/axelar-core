package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/funcs"
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

func (v voteHandler) HandleFailedPoll(ctx sdk.Context, poll vote.Poll) error {
	md := mustGetMetadata(poll)
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.PollFailed{
		TxID:   md.TxID,
		Chain:  md.Chain,
		PollID: poll.GetID(),
	}))

	return nil
}

func (v voteHandler) IsFalsyResult(result codec.ProtoMarshaler) bool {
	return len(result.(*types.VoteEvents).Events) == 0
}

func (v voteHandler) HandleExpiredPoll(ctx sdk.Context, poll vote.Poll) error {
	rewardPoolName, ok := poll.GetRewardPoolName()
	if !ok {
		return fmt.Errorf("reward pool not set for poll %s", poll.GetID().String())
	}

	// TODO: MarkMissingVote for those who didn't vote in time. Need
	// to be able to get chain of expired polls in order to do it.
	rewardPool := v.rewarder.GetPool(ctx, rewardPoolName)
	// Penalize voters who failed to vote
	for _, voter := range poll.GetVoters() {
		if !poll.HasVoted(voter) {
			rewardPool.ClearRewards(voter)
			v.keeper.Logger(ctx).Debug(fmt.Sprintf("penalized voter %s due to timeout", voter.String()),
				"voter", voter.String(),
				"poll", poll.GetID().String())
		}
	}

	md := mustGetMetadata(poll)
	funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.PollExpired{
		TxID:   md.TxID,
		Chain:  md.Chain,
		PollID: poll.GetID(),
	}))

	return nil
}

func (v voteHandler) HandleCompletedPoll(ctx sdk.Context, poll vote.Poll) error {
	voteEvents := poll.GetResult().(*types.VoteEvents)

	chain, ok := v.nexus.GetChain(ctx, voteEvents.Chain)
	if !ok {
		return fmt.Errorf("%s is not a registered chain", voteEvents.Chain)
	}

	rewardPoolName, ok := poll.GetRewardPoolName()
	if !ok {
		return fmt.Errorf("reward pool not set for poll %s", poll.GetID().String())
	}

	rewardPool := v.rewarder.GetPool(ctx, rewardPoolName)

	chainState := v.nexus.GetChainState(ctx, chain)
	for _, voter := range poll.GetVoters() {
		hasVoted := poll.HasVoted(voter)
		hasVotedIncorrectly := hasVoted && !poll.HasVotedCorrectly(voter)

		chainState.MarkMissingVote(voter, !hasVoted)
		chainState.MarkIncorrectVote(voter, hasVotedIncorrectly)

		v.keeper.Logger(ctx).Debug(fmt.Sprintf("marked voter %s behaviour", voter.String()),
			"voter", voter.String(),
			"missing_vote", !hasVoted,
			"incorrect_vote", hasVotedIncorrectly,
		)

		switch {
		case hasVotedIncorrectly, !hasVoted:
			rewardPool.ClearRewards(voter)
			v.keeper.Logger(ctx).Debug(fmt.Sprintf("penalized voter %s due to incorrect vote or missing vote", voter.String()),
				"voter", voter.String(),
				"poll", poll.GetID().String())
		default:
			if err := rewardPool.ReleaseRewards(voter); err != nil {
				return err
			}
			v.keeper.Logger(ctx).Debug(fmt.Sprintf("released rewards for voter %s", voter.String()),
				"voter", voter.String(),
				"poll", poll.GetID().String())
		}
	}
	v.nexus.SetChainState(ctx, chainState)

	if v.IsFalsyResult(voteEvents) {
		md := mustGetMetadata(poll)
		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(&types.NoEventsConfirmed{
			TxID:   md.TxID,
			Chain:  md.Chain,
			PollID: poll.GetID(),
		}))
	}

	return nil
}

func (v voteHandler) HandleResult(ctx sdk.Context, result codec.ProtoMarshaler) error {
	voteEvents := result.(*types.VoteEvents)

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

	ck := v.keeper.ForChain(chain.Name)
	for _, event := range voteEvents.Events {
		if err := handleEvent(ctx, ck, event, chain); err != nil {
			return err
		}
	}

	return nil
}

func handleEvent(ctx sdk.Context, ck types.ChainKeeper, event types.Event, chain nexus.Chain) error {
	// validate event
	// TODO: move to ValidateBasic of msg_vote
	if err := event.ValidateBasic(); err != nil {
		return fmt.Errorf("event %s: %s", event.GetID(), err.Error())
	}

	// check if event confirmed before
	eventID := event.GetID()
	if _, ok := ck.GetEvent(ctx, eventID); ok {
		return fmt.Errorf("event %s is already confirmed", eventID)
	}
	if err := ck.SetConfirmedEvent(ctx, event); err != nil {
		panic(err)
	}
	ck.Logger(ctx).Info(fmt.Sprintf("confirmed %s event %s in transaction %s", chain.Name, eventID, event.TxID.Hex()))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeEventConfirmation,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyChain, event.Chain.String()),
			sdk.NewAttribute(types.AttributeKeyTxID, event.TxID.Hex()),
			sdk.NewAttribute(types.AttributeKeyEventID, string(event.GetID())),
			sdk.NewAttribute(types.AttributeKeyEventType, event.GetEventType()),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueConfirm)),
	)

	return nil
}

func mustGetMetadata(poll vote.Poll) types.PollMetadata {
	md := funcs.MustOk(poll.GetMetaData())
	chainTxID, ok := md.(*types.PollMetadata)
	if !ok {
		panic(fmt.Sprintf("poll metadata should be of type %T", &types.PollMetadata{}))
	}
	return *chainTxID
}
