/*
This package manages second layer voting. It caches votes until they are sent out in a batch and tallies the results.
*/
package keeper

import (
	"crypto/sha256"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/utils"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

const (
	pendingVotes       = "pending"
	votingIntervalKey  = "votingInterval"
	votingThresholdKey = "votingThreshold"
	pollPrefix         = "poll_"
	talliedPrefix      = "tallied_"
	addrPrefix         = "addr_"

	// Dummy values: the values do not matter, used as markers
	voted         byte = 0
	indexNotFound      = -1
)

type Keeper struct {
	subjectiveStore store.SubjectiveStore
	storeKey        sdk.StoreKey
	cdc             *codec.Codec
	broadcaster     types.Broadcaster
	snapshotter     types.Snapshotter
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, subjectiveStore store.SubjectiveStore, snapshotter types.Snapshotter, broadcaster types.Broadcaster) Keeper {
	keeper := Keeper{
		subjectiveStore: subjectiveStore,
		storeKey:        key,
		cdc:             cdc,
		broadcaster:     broadcaster,
		snapshotter:     snapshotter,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetVotingInterval sets the interval in which votes are supposed to be broadcast
func (k Keeper) SetVotingInterval(ctx sdk.Context, votingInterval int64) {
	ctx.KVStore(k.storeKey).Set([]byte(votingIntervalKey), k.cdc.MustMarshalBinaryLengthPrefixed(votingInterval))
}

// GetVotingInterval returns the interval in which votes are supposed to be broadcast
func (k Keeper) GetVotingInterval(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(votingIntervalKey))

	var interval int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &interval)
	return interval
}

// SetVotingThreshold sets the voting power threshold that must be reached to decide a poll
func (k Keeper) SetVotingThreshold(ctx sdk.Context, threshold utils.Threshold) {
	ctx.KVStore(k.storeKey).Set([]byte(votingThresholdKey), k.cdc.MustMarshalBinaryLengthPrefixed(threshold))
}

// GetVotingThreshold returns the voting power threshold that must be reached to decide a poll
func (k Keeper) GetVotingThreshold(ctx sdk.Context) utils.Threshold {
	rawThreshold := ctx.KVStore(k.storeKey).Get([]byte(votingThresholdKey))
	var threshold utils.Threshold
	k.cdc.MustUnmarshalBinaryLengthPrefixed(rawThreshold, &threshold)
	return threshold
}

// InitPoll initializes a new poll. This is the first step of the voting protocol.
// The Keeper only accepts votes for initialized polls.
func (k Keeper) InitPoll(ctx sdk.Context, poll exported.PollMeta) error {
	if k.getPoll(ctx, poll) != nil {
		return fmt.Errorf("poll with same name already exists")
	}

	r := k.snapshotter.GetLatestRound(ctx)
	k.setPoll(ctx, types.Poll{Meta: poll, ValidatorSnapshotRound: r})
	return nil
}

// DeletePoll deletes the specified poll.
func (k Keeper) DeletePoll(ctx sdk.Context, poll exported.PollMeta) {
	// delete poll
	ctx.KVStore(k.storeKey).Delete([]byte(pollPrefix + poll.String()))

	// delete tallied votes index for poll
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(talliedPrefix+poll.String()))
	for ; iter.Valid(); iter.Next() {
		ctx.KVStore(k.storeKey).Delete(iter.Key())
	}

	// delete voter index for poll
	iter = sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(addrPrefix+poll.String()))
	for ; iter.Valid(); iter.Next() {
		ctx.KVStore(k.storeKey).Delete(iter.Key())
	}
}

// RecordVote readies a vote to be broadcast to the entire network.
// Votes are only valid if they correspond to a previously initialized poll.
// Depending on the voting interval, multiple votes might be batched together when broadcasting.
func (k Keeper) RecordVote(ctx sdk.Context, vote exported.MsgVote) error {
	if k.getPoll(ctx, vote.Poll()) == nil {
		return fmt.Errorf("no poll registered with the given id")
	}

	votes := k.getPendingVotes()
	for _, existingVote := range votes {
		if existingVote.Poll() == vote.Poll() {
			return fmt.Errorf(fmt.Sprintf("already recorded a vote for poll %s", vote.Poll()))
		}
	}
	votes = append(votes, vote)
	k.Logger(ctx).Debug(fmt.Sprintf("new vote for poll %s, data hash: %s", vote.Poll().String(), k.hash(vote.Data())))
	k.setPendingVotes(votes)

	return nil
}

// SendVotes broadcasts all unpublished votes to the entire network.
func (k Keeper) SendVotes(ctx sdk.Context) {
	votes := k.getPendingVotes()
	k.Logger(ctx).Debug(fmt.Sprintf("unpublished votes:%v", len(votes)))

	if len(votes) == 0 {
		return
	}

	// Reset votes for the next round
	k.setPendingVotes(nil)

	// Broadcast is a subjective action, so it must not cause non-deterministic changes to the tx execution.
	// Because of this and to prevent a deadlock it needs to run in its own goroutine without any callbacks.
	go func(logger log.Logger) {
		var msgs []broadcast.MsgWithSenderSetter
		for _, vote := range votes {
			msgs = append(msgs, vote)
		}

		err := k.broadcaster.Broadcast(ctx, msgs)
		if err != nil {
			logger.Error(sdkerrors.Wrap(err, "broadcasting votes failed").Error())
		} else {
			logger.Debug("broadcasting votes")
		}
	}(k.Logger(ctx))
}

// TallyVote tallies votes that have been broadcast. Each validator can only vote once per poll.
func (k Keeper) TallyVote(ctx sdk.Context, vote exported.MsgVote) error {
	poll := k.getPoll(ctx, vote.Poll())
	if poll == nil {
		return fmt.Errorf("poll does not exist or is closed")
	}

	valAddress := k.broadcaster.GetPrincipal(ctx, vote.GetSigners()[0])
	if valAddress == nil {
		err := fmt.Errorf("account %v is not registered as a validator proxy", vote.GetSigners()[0])
		return err
	}

	snap, ok := k.snapshotter.GetSnapshot(ctx, poll.ValidatorSnapshotRound)
	if !ok {
		return fmt.Errorf("no snapshot found for round %d", poll.ValidatorSnapshotRound)
	}

	validator, ok := snap.GetValidator(valAddress)
	if !ok {
		return fmt.Errorf("address %s is not eligible to vote in this poll", valAddress.String())
	}

	if k.getHasVoted(ctx, vote.Poll(), valAddress) {
		return fmt.Errorf("each validator can only vote once")
	}

	// if the poll is already decided there is no need to keep track of further votes
	if poll.Result != nil {
		return nil
	}

	k.setHasVoted(ctx, vote.Poll(), valAddress)
	var talliedVote types.TalliedVote
	// check if others match this vote, create a new unique entry if not, simply add voting power if match is found
	i := k.getTalliedVoteIdx(ctx, vote)
	if i == indexNotFound {
		talliedVote = types.TalliedVote{
			Tally: sdk.NewInt(validator.GetConsensusPower()),
			Data:  vote.Data(),
		}

		poll.Votes = append(poll.Votes, talliedVote)
		k.setTalliedVoteIdx(ctx, vote, len(poll.Votes)-1)
	} else {
		// this assignment copies the value, so we need to write it back into the array
		talliedVote = poll.Votes[i]
		talliedVote.Tally = talliedVote.Tally.AddRaw(validator.GetConsensusPower())
		poll.Votes[i] = talliedVote
	}

	threshold := k.GetVotingThreshold(ctx)
	if threshold.IsMet(talliedVote.Tally, snap.TotalPower) {
		k.Logger(ctx).Debug(fmt.Sprintf("threshold of %d/%d has been met for %s: %s/%s",
			threshold.Numerator, threshold.Denominator, vote.Poll(), talliedVote.Tally.String(), snap.TotalPower.String()))
		poll.Result = talliedVote.Data
	}

	k.setPoll(ctx, *poll)
	return nil
}

// Result returns the decided outcome of a poll. Returns nil if the poll is still undecided or does not exist.
func (k Keeper) Result(ctx sdk.Context, pollMeta exported.PollMeta) exported.VotingData {
	// This unmarshals all votes for this poll, which is not needed in this context.
	// Should it become a performance concern we could split the result off into a separate data structure
	poll := k.getPoll(ctx, pollMeta)
	if poll == nil {
		return nil
	}
	return poll.Result
}

// Because votes may differ between nodes they need to be stored outside the regular kvstore
// (whose hash becomes part of the Merkle tree)
func (k Keeper) getPendingVotes() []exported.MsgVote {
	bz := k.subjectiveStore.Get([]byte(pendingVotes))
	if bz == nil {
		return nil
	}
	var votes []exported.MsgVote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &votes)
	return votes
}

// See getPendingVotes
func (k Keeper) setPendingVotes(votes []exported.MsgVote) {
	k.subjectiveStore.Set([]byte(pendingVotes), k.cdc.MustMarshalBinaryLengthPrefixed(votes))
}

// using a pointer reference to adhere to the pattern of returning nil if value is not found
func (k Keeper) getPoll(ctx sdk.Context, pollMeta exported.PollMeta) *types.Poll {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pollPrefix + pollMeta.String()))
	if bz == nil {
		return nil
	}

	var poll types.Poll
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &poll)
	return &poll
}

func (k Keeper) setPoll(ctx sdk.Context, poll types.Poll) {
	ctx.KVStore(k.storeKey).Set([]byte(pollPrefix+poll.Meta.String()), k.cdc.MustMarshalBinaryLengthPrefixed(poll))
}

// To adhere to the same one-return-value pattern as the other getters return a marker value if not found
// (returning an int with a pointer reference to be able to return nil instead seems bizarre)
func (k Keeper) getTalliedVoteIdx(ctx sdk.Context, vote exported.MsgVote) int {
	// check if there have been identical votes
	bz := ctx.KVStore(k.storeKey).Get(k.talliedVoteKey(vote))
	if bz == nil {
		return indexNotFound
	}
	var i int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &i)
	return i
}

func (k Keeper) setTalliedVoteIdx(ctx sdk.Context, vote exported.MsgVote, i int) {
	voteKey := k.talliedVoteKey(vote)
	ctx.KVStore(k.storeKey).Set(voteKey, k.cdc.MustMarshalBinaryLengthPrefixed(i))
}

func (k Keeper) getHasVoted(ctx sdk.Context, poll exported.PollMeta, address sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(addrPrefix + poll.String() + address.String()))
}

func (k Keeper) setHasVoted(ctx sdk.Context, poll exported.PollMeta, address sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(addrPrefix+poll.String()+address.String()), []byte{voted})
}

func (k Keeper) talliedVoteKey(vote exported.MsgVote) []byte {
	return []byte(talliedPrefix + vote.Poll().String() + k.hash(vote.Data()))
}

func (k Keeper) hash(data exported.VotingData) string {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(data)
	h := sha256.Sum256(bz)
	return string(h[:])
}
