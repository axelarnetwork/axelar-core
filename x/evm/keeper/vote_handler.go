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

// NewVoteHandler returns the handler for processing vote delivered by the vote module
func NewVoteHandler(cdc codec.Codec, keeper types.BaseKeeper, nexus types.Nexus) vote.VoteHandler {
	return func(ctx sdk.Context, poll vote.Poll) error {
		voteEvents, err := types.UnpackEvents(cdc, poll.GetResult().(*vote.Vote).Result)
		if err != nil {
			return err
		}

		if slices.Any(voteEvents.Events, func(event types.Event) bool { return event.Chain != voteEvents.Chain }) {
			return fmt.Errorf("events are not from the same source chain")
		}

		chain, ok := nexus.GetChain(ctx, voteEvents.Chain)
		if !ok {
			return fmt.Errorf("%s is not a registered chain", voteEvents.Chain)
		}

		if !keeper.HasChain(ctx, voteEvents.Chain) {
			return fmt.Errorf("%s is not an evm chain", voteEvents.Chain)
		}

		for _, voter := range poll.GetVoters() {
			hasVoted := poll.HasVoted(voter.Validator)
			hasVotedIncorrectly := hasVoted && !poll.HasVotedCorrectly(voter.Validator)

			nexus.MarkChainMaintainerMissingVote(ctx, chain, voter.Validator, !hasVoted)
			nexus.MarkChainMaintainerIncorrectVote(ctx, chain, voter.Validator, hasVotedIncorrectly)
		}

		if len(voteEvents.Events) == 0 {
			poll.AllowOverride()

			return nil
		}

		chainK := keeper.ForChain(chain.Name)
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
}

func handleEvents(ctx sdk.Context, ck types.ChainKeeper, events []types.Event, chain nexus.Chain) error {
	for _, event := range events {
		var err error
		// validate event
		err = event.ValidateBasic()
		if err != nil {
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
