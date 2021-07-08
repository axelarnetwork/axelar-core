/*
Package keeper manages second layer voting. It caches votes until they are sent out in a batch and tallies the results.
*/
package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

var (
	thresholdKey = utils.KeyFromStr("votingThreshold")
	pollPrefix   = utils.KeyFromStr("poll")
	votesPrefix  = utils.KeyFromStr("votes")
)

// Keeper - the vote module's keeper
type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         codec.BinaryMarshaler
	snapshotter types.Snapshotter
}

// NewKeeper - keeper constructor
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, snapshotter types.Snapshotter) Keeper {
	keeper := Keeper{
		storeKey:    key,
		cdc:         cdc,
		snapshotter: snapshotter,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetDefaultVotingThreshold sets the default voting power threshold that must be reached to decide a poll
func (k Keeper) SetDefaultVotingThreshold(ctx sdk.Context, threshold utils.Threshold) {
	k.getStore(ctx).Set(thresholdKey, &threshold)
}

// GetDefaultVotingThreshold returns the default voting power threshold that must be reached to decide a poll
func (k Keeper) GetDefaultVotingThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.getStore(ctx).Get(thresholdKey, &threshold)

	return threshold
}

// InitPoll initializes a new poll. This is the first step of the voting protocol.
// The Keeper only accepts votes for initialized polls.
func (k Keeper) InitPoll(ctx sdk.Context, pollKey exported.PollKey, snapshotSeqNo int64, expiresAt int64, threshold ...utils.Threshold) error {
	poll := k.GetPollMetadata(ctx, pollKey)

	switch {
	case poll.Is(types.Pending):
		return fmt.Errorf("poll %s already exists and has not expired yet", pollKey.String())
	case poll.Is(types.Completed):
		return fmt.Errorf("poll %s already exists and has a result", pollKey.String())
	case !poll.Is(types.NonExistant):
		k.Logger(ctx).Debug(fmt.Sprintf("deleting poll %s due to expiry", pollKey.String()))
		k.deletePoll(ctx, pollKey)
	}

	t := k.GetDefaultVotingThreshold(ctx)
	if len(threshold) > 0 {
		t = threshold[0]
	}
	k.setPollMetadata(ctx, types.NewPollMetaData(pollKey, snapshotSeqNo, expiresAt, t))
	return nil
}

// TallyVote tallies votes that have been broadcast. Each validator can only vote once per poll.
func (k Keeper) TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollKey exported.PollKey, data codec.ProtoMarshaler) (types.PollMetadata, error) {
	valAddress := k.snapshotter.GetPrincipal(ctx, sender)
	if valAddress == nil {
		return types.PollMetadata{}, fmt.Errorf("account %v is not registered as a validator proxy", sender.String())
	}

	poll := k.getPoll(ctx, pollKey)
	if poll.Is(types.NonExistant) {
		return types.PollMetadata{}, fmt.Errorf("poll does not exist")
	}

	// if the poll is already decided there is no need to keep track of further votes
	if poll.Is(types.Completed) || poll.Is(types.Failed) {
		return poll.PollMetadata, nil
	}

	if err := poll.Vote(valAddress, data); err != nil {
		return types.PollMetadata{}, err
	}

	switch {
	case poll.Is(types.Completed):
		k.Logger(ctx).Debug(fmt.Sprintf("poll %s (threshold: %d/%d) completed", pollKey,
			poll.VotingThreshold.Numerator, poll.VotingThreshold.Denominator))
	case poll.Is(types.Failed):
		k.Logger(ctx).Debug(fmt.Sprintf("poll %s (threshold: %d/%d) failed, voters could not agree on single value", pollKey,
			poll.VotingThreshold.Numerator, poll.VotingThreshold.Denominator))
	}

	k.setPoll(ctx, poll)

	return poll.PollMetadata, nil
}

// GetPollMetadata returns the poll given poll metadata
func (k Keeper) GetPollMetadata(ctx sdk.Context, pollKey exported.PollKey) types.PollMetadata {
	var poll types.PollMetadata
	if ok := k.getStore(ctx).Get(pollPrefix.AppendStr(pollKey.String()), &poll); !ok {
		return types.PollMetadata{State: types.NonExistant}
	}

	return poll.UpdateBlockHeight(ctx.BlockHeight())
}

func (k Keeper) deletePoll(ctx sdk.Context, pollKey exported.PollKey) {
	// delete poll
	k.getStore(ctx).Delete(pollPrefix.AppendStr(pollKey.String()))

	// delete tallied votes index for poll
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), votesPrefix.AppendStr(pollKey.String()).AsKey())
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		k.getStore(ctx).Delete(utils.KeyFromBz(iter.Key()))
	}
}

func (k Keeper) setPollMetadata(ctx sdk.Context, poll types.PollMetadata) {
	k.getStore(ctx).Set(pollPrefix.AppendStr(poll.Key.String()), &poll)
}

func (k Keeper) getPoll(ctx sdk.Context, key exported.PollKey) types.Poll {
	metadata := k.GetPollMetadata(ctx, key)
	snap, _ := k.snapshotter.GetSnapshot(ctx, metadata.SnapshotSeqNo)
	return types.NewPoll(metadata, k.getVotes(ctx, key), snap)
}

func (k Keeper) getVotes(ctx sdk.Context, key exported.PollKey) []types.TalliedVote {
	var votes []types.TalliedVote
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), votesPrefix.AppendStr(key.String()).AsKey())
	defer utils.CloseLogError(iter, k.Logger(ctx))
	for ; iter.Valid(); iter.Next() {
		var vote types.TalliedVote
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &vote)
		votes = append(votes, vote)
	}
	return votes
}

func (k Keeper) setPoll(ctx sdk.Context, poll types.Poll) {
	k.setPollMetadata(ctx, poll.PollMetadata)
	k.setVotes(ctx, poll.Key, poll.Votes)
}

func (k Keeper) setVotes(ctx sdk.Context, key exported.PollKey, votes []types.FlaggedVote) {
	for i, vote := range votes {
		if vote.Dirty {
			k.getStore(ctx).Set(votesPrefix.AppendStr(key.String()).AppendStr(strconv.Itoa(i)), &vote.Vote)
		}
	}
}

func (k Keeper) getStore(ctx sdk.Context) utils.NormalizedKVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
