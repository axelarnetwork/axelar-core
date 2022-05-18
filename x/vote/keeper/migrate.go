package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	exported0_17 "github.com/axelarnetwork/axelar-core/x/vote/exported017"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	types0_17 "github.com/axelarnetwork/axelar-core/x/vote/types-0.17"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.17 to v0.18. The
// migration includes:
// - delete all pending polls
// - migrate all completed polls
func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		err := migrateVotes(ctx, k)
		if err != nil {
			return err
		}

		deleteAllPendingPolls(ctx, k)
		migrateAllCompletedPolls(ctx, k)

		return nil
	}
}

func migrateVotes(ctx sdk.Context, k Keeper) error {
	metadatas := k.getPollMetadatasOld(ctx)
	for _, metadata := range metadatas {
		newMetadata := exported.PollMetadata{
			Key:              metadata.Key,
			ExpiresAt:        metadata.ExpiresAt,
			Result:           MigrateVoteData(k.cdc, nexus.ChainName(metadata.RewardPoolName), metadata.Result, k.Logger(ctx)),
			VotingThreshold:  metadata.VotingThreshold,
			State:            metadata.State,
			MinVoterCount:    metadata.MinVoterCount,
			Voters:           metadata.Voters,
			TotalVotingPower: metadata.TotalVotingPower,
			RewardPoolName:   metadata.RewardPoolName,
		}

		pollStore := k.newPollStore(ctx, newMetadata.Key)
		pollStore.SetMetadata(newMetadata)
		votes := pollStore.GetVotesOld()
		for _, vote := range votes {
			newVote := types.TalliedVote{
				Tally:  vote.Tally,
				Voters: vote.Voters,
				Data:   MigrateVoteData(k.cdc, nexus.ChainName(metadata.RewardPoolName), vote.Data, k.Logger(ctx)),
			}
			pollStore.Set(votesPrefix.AppendStr(pollStore.key.String()).AppendStr(hash(newVote.Data)), &newVote)
		}
	}

	return assertMigrationSuccessful(ctx, k)
}

func assertMigrationSuccessful(ctx sdk.Context, k Keeper) error {
	metadatas := k.getPollMetadatas(ctx)
	for _, metadata := range metadatas {
		if metadata.Result != nil && metadata.Result.GetCachedValue() == nil {
			return fmt.Errorf("failed to verify poll result for %s", hex.EncodeToString(k.cdc.MustMarshalLengthPrefixed(&metadata)))
		}

		pollStore := k.newPollStore(ctx, metadata.Key)
		votes := pollStore.GetVotes()
		for _, vote := range votes {
			if vote.Data.GetCachedValue() == nil {
				return fmt.Errorf("failed to verify tallied vote data for %s", hex.EncodeToString(k.cdc.MustMarshalLengthPrefixed(&vote)))
			}
		}
	}
	return nil
}

// MigrateVoteData migrates vote results from an Any slice to a single Any value
func MigrateVoteData(cdc codec.BinaryCodec, chain nexus.ChainName, data *codectypes.Any, logger log.Logger) *codectypes.Any {
	if data == nil {
		return nil
	}

	switch d := data.GetCachedValue().(type) {
	case *exported0_17.Vote:
		events, err := unpackEvents(cdc, d.Results)
		if err != nil {
			logger.Error("failed tp unpack vote results for %s", hex.EncodeToString(cdc.MustMarshalLengthPrefixed(d)))
			return nil
		}

		if len(events) != 0 {
			chain = events[0].Chain
		}
		result, err := evmtypes.PackEvents(chain, events)
		if err != nil {
			logger.Error("failed to pack events for %s", hex.EncodeToString(cdc.MustMarshalLengthPrefixed(d)))
			return nil
		}
		packedVote, err := codectypes.NewAnyWithValue(&exported.Vote{Result: result})
		if err != nil {
			logger.Error("failed to pack vote for %s", hex.EncodeToString(cdc.MustMarshalLengthPrefixed(d)))
			return nil
		}
		return packedVote
	default:
		// in all other cases the data struct stays the same
		logger.Info(fmt.Sprintf("ignoring migration for data of type %s", data.GoString()))
		return data
	}
}

func (k Keeper) getPollMetadatasOld(ctx sdk.Context) []exported0_17.PollMetadata {
	var pollMetadatas []exported0_17.PollMetadata

	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pollMetadata exported0_17.PollMetadata
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &pollMetadata)
		pollMetadatas = append(pollMetadatas, pollMetadata)
	}

	return pollMetadatas
}

func (p *pollStore) GetVotesOld() []types0_17.TalliedVote {
	iter := p.Iterator(votesPrefix.AppendStr(p.key.String()))
	defer utils.CloseLogError(iter, p.logger)

	var votes []types0_17.TalliedVote
	for ; iter.Valid(); iter.Next() {
		var vote types0_17.TalliedVote
		iter.UnmarshalValue(&vote)
		votes = append(votes, vote)
	}
	return votes
}

// UnpackEvents converts Any slice to Events
func unpackEvents(cdc codec.BinaryCodec, eventsAny []*codectypes.Any) ([]evmtypes.Event, error) {
	var events []evmtypes.Event
	for _, e := range eventsAny {
		var event evmtypes.Event
		if err := cdc.Unmarshal(e.Value, &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func deleteAllPendingPolls(ctx sdk.Context, k Keeper) {
	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pollMetadata exported.PollMetadata
		iter.UnmarshalValue(&pollMetadata)

		if !pollMetadata.Is(exported.Pending) {
			continue
		}

		k.newPollStore(ctx, pollMetadata.Key).DeletePoll()
	}
}

func migrateAllCompletedPolls(ctx sdk.Context, k Keeper) {
	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pollMetadata exported.PollMetadata
		iter.UnmarshalValue(&pollMetadata)

		if !pollMetadata.Is(exported.Completed) {
			continue
		}

		poll := k.newPollStore(ctx, pollMetadata.Key)
		voterIter := k.getKVStore(ctx).Iterator(voterPrefix.AppendStr(poll.key.String()))
		for ; voterIter.Valid(); voterIter.Next() {
			poll.KVStore.Set(voterIter.GetKey(), &types.VoteRecord{IsLate: false})
		}
		utils.CloseLogError(voterIter, k.Logger(ctx))
		// The actual completed at cannot be retrieved anymore, but need to
		// make it valid
		pollMetadata.CompletedAt = 1
		poll.SetMetadata(pollMetadata)
	}
}
