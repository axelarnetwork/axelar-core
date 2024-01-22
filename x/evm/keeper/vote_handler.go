package keeper

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/funcs"
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
	events.Emit(ctx, &types.PollFailed{
		TxID:   md.TxID,
		Chain:  md.Chain,
		PollID: poll.GetID(),
	})

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

	md := mustGetMetadata(poll)
	rewardPool := v.rewarder.GetPool(ctx, rewardPoolName)
	chain, ok := v.nexus.GetChain(ctx, md.Chain)
	if !ok {
		return fmt.Errorf("%s is not a registered chain", md.Chain)
	}
	// Penalize voters who failed to vote
	for _, voter := range poll.GetVoters() {
		hasVoted := poll.HasVoted(voter)
		if maintainerState, ok := v.nexus.GetChainMaintainerState(ctx, chain, voter); ok {
			maintainerState.MarkMissingVote(!hasVoted)
			funcs.MustNoErr(v.nexus.SetChainMaintainerState(ctx, maintainerState))

			v.keeper.Logger(ctx).Debug(fmt.Sprintf("marked voter %s behaviour", voter.String()),
				"voter", voter.String(),
				"missing_vote", !hasVoted,
				"poll", poll.GetID().String(),
			)
		}

		if !hasVoted {
			rewardPool.ClearRewards(voter)
			v.keeper.Logger(ctx).Debug(fmt.Sprintf("penalized voter %s due to timeout", voter.String()),
				"voter", voter.String(),
				"poll", poll.GetID().String())
		}
	}

	events.Emit(ctx, &types.PollExpired{
		TxID:   md.TxID,
		Chain:  md.Chain,
		PollID: poll.GetID(),
	})

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

	for _, voter := range poll.GetVoters() {
		maintainerState, ok := v.nexus.GetChainMaintainerState(ctx, chain, voter)
		if !ok {
			continue // voter is no longer a chain maintainer, so recording the state is irrelevant
		}

		hasVoted := poll.HasVoted(voter)
		hasVotedIncorrectly := hasVoted && !poll.HasVotedCorrectly(voter)

		maintainerState.MarkMissingVote(!hasVoted)
		maintainerState.MarkIncorrectVote(hasVotedIncorrectly)
		funcs.MustNoErr(v.nexus.SetChainMaintainerState(ctx, maintainerState))

		v.keeper.Logger(ctx).Debug(fmt.Sprintf("marked voter %s behaviour", voter.String()),
			"voter", voter.String(),
			"missing_vote", !hasVoted,
			"incorrect_vote", hasVotedIncorrectly,
			"poll", poll.GetID().String(),
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

	md := mustGetMetadata(poll)
	if v.IsFalsyResult(voteEvents) {
		events.Emit(ctx, &types.NoEventsConfirmed{
			TxID:   md.TxID,
			Chain:  md.Chain,
			PollID: poll.GetID(),
		})
	}

	events.Emit(ctx, &types.PollCompleted{
		TxID:   md.TxID,
		Chain:  md.Chain,
		PollID: poll.GetID(),
	})

	return nil
}

func (v voteHandler) HandleResult(ctx sdk.Context, result codec.ProtoMarshaler) error {
	voteEvents := result.(*types.VoteEvents)

	if v.IsFalsyResult(result) {
		return nil
	}

	chain, ok := v.nexus.GetChain(ctx, voteEvents.Chain)
	if !ok {
		return fmt.Errorf("%s is not a registered chain", voteEvents.Chain)
	}

	ck, err := v.keeper.ForChain(ctx, chain.Name)
	if err != nil {
		return fmt.Errorf("%s is not an evm chain", voteEvents.Chain)
	}

	for _, event := range voteEvents.Events {
		if err := v.handleEvent(ctx, ck, event, chain); err != nil {
			return err
		}
	}

	return nil
}

func (v voteHandler) handleEvent(ctx sdk.Context, ck types.ChainKeeper, event types.Event, chain nexus.Chain) error {
	if err := ck.SetConfirmedEvent(ctx, event); err != nil {
		return err
	}

	// Event_ContractCall is no longer directly handled by the EVM module,
	// which bypassed nexus routing
	switch event.GetEvent().(type) {
	case *types.Event_ContractCall:
		if err := v.handleContractCall(ctx, ck, event); err != nil {
			return err
		}
	default:
		funcs.MustNoErr(ck.EnqueueConfirmedEvent(ctx, event.GetID()))
	}

	ck.Logger(ctx).Info(fmt.Sprintf("confirmed %s event %s in transaction %s", chain.Name, event.GetID(), event.TxID.Hex()))

	// Deprecated
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

func (v voteHandler) handleContractCall(ctx sdk.Context, ck types.ChainKeeper, event types.Event) error {
	msg := mustToGeneralMessage(ctx, v.nexus, event)

	if err := v.nexus.SetNewMessage(ctx, msg); err != nil {
		return err
	}

	funcs.MustNoErr(v.nexus.EnqueueRouteMessage(ctx, msg.ID))
	funcs.MustNoErr(ck.SetEventCompleted(ctx, event.GetID()))

	return nil
}

func mustToGeneralMessage(ctx sdk.Context, n types.Nexus, event types.Event) nexus.GeneralMessage {
	id := string(event.GetID())
	contractCall := event.GetEvent().(*types.Event_ContractCall).ContractCall

	sourceChain := funcs.MustOk(n.GetChain(ctx, event.Chain))
	sender := nexus.CrossChainAddress{Chain: sourceChain, Address: contractCall.Sender.Hex()}

	destinationChain, ok := n.GetChain(ctx, contractCall.DestinationChain)
	if !ok {
		// try forwarding it to wasm router if destination chain is not registered
		destinationChain = nexus.Chain{Name: contractCall.DestinationChain, SupportsForeignAssets: false, KeyType: tss.None, Module: wasm.ModuleName}
	}
	recipient := nexus.CrossChainAddress{Chain: destinationChain, Address: contractCall.ContractAddress}

	return nexus.NewGeneralMessage(id, sender, recipient, contractCall.PayloadHash.Bytes(), event.TxID.Bytes(), event.Index, nil)
}

func mustGetMetadata(poll vote.Poll) types.PollMetadata {
	md := funcs.MustOk(poll.GetMetaData())
	metadata, ok := md.(*types.PollMetadata)
	if !ok {
		panic(fmt.Sprintf("poll metadata should be of type %T", &types.PollMetadata{}))
	}
	return *metadata
}
