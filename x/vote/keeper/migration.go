package keeper

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported2"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types2"
)

func GetMigrationHandler(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		migrateVotes(ctx, k)
		return nil
	}
}

func migrateVotes(ctx sdk.Context, k Keeper) {
	metadatas := k.getPollMetadatasOld(ctx)
	for _, metadata := range metadatas {
		newMetadata := exported.PollMetadata{
			Key:              metadata.Key,
			ExpiresAt:        metadata.ExpiresAt,
			Result:           MigrateVoteData(k.cdc, metadata.Result, k.Logger(ctx)),
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
				Data:   MigrateVoteData(k.cdc, vote.Data, k.Logger(ctx)),
			}
			pollStore.Set(votesPrefix.AppendStr(pollStore.key.String()).AppendStr(newVote.Hash()), &newVote)
		}
	}
}

func MigrateVoteData(cdc codec.BinaryCodec, data *codectypes.Any, logger log.Logger) *codectypes.Any {
	switch d := data.GetCachedValue().(type) {
	case *exported2.Vote:
		events, err := unpackEvents(cdc, d.Results)
		if err != nil {
			logger.Error("failed tp unpack vote results for %s", hex.EncodeToString(cdc.MustMarshalLengthPrefixed(d)))
			return nil
		}

		result, err := evmtypes.PackEvents(events[0].Chain, events)
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
		// in all other cases the data struct stayed the same
		return data
	}
}

func (k Keeper) getPollMetadatasOld(ctx sdk.Context) []exported2.PollMetadata {
	var pollMetadatas []exported2.PollMetadata

	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pollMetadata exported2.PollMetadata
		k.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &pollMetadata)
		pollMetadatas = append(pollMetadatas, pollMetadata)
	}

	return pollMetadatas
}

func (p *pollStore) GetVotesOld() []types2.TalliedVote {
	iter := p.Iterator(votesPrefix.AppendStr(p.key.String()))
	defer utils.CloseLogError(iter, p.logger)

	var votes []types2.TalliedVote
	for ; iter.Valid(); iter.Next() {
		var vote types2.TalliedVote
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
