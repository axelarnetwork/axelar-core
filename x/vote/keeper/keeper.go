/*
Package keeper manages second layer voting. It caches votes until they are sent out in a batch and tallies the results.
*/
package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

const (
	votingIntervalKey  = "votingInterval"
	votingThresholdKey = "votingThreshold"
	pollPrefix         = "poll_"
	talliedPrefix      = "tallied_"
	addrPrefix         = "addr_"

	// Dummy values: the values do not matter, used as markers
	voted         byte = 0
	indexNotFound int  = -1
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

// SetVotingInterval sets the interval in which votes are supposed to be broadcast
func (k Keeper) SetVotingInterval(ctx sdk.Context, votingInterval int64) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(votingInterval))

	ctx.KVStore(k.storeKey).Set([]byte(votingIntervalKey), bz)
}

// GetVotingInterval returns the interval in which votes are supposed to be broadcast
func (k Keeper) GetVotingInterval(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(votingIntervalKey))

	return int64(binary.LittleEndian.Uint64(bz))
}

// SetVotingThreshold sets the voting power threshold that must be reached to decide a poll
func (k Keeper) SetVotingThreshold(ctx sdk.Context, threshold utils.Threshold) {
	ctx.KVStore(k.storeKey).Set([]byte(votingThresholdKey), k.cdc.MustMarshalBinaryLengthPrefixed(&threshold))
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
func (k Keeper) InitPoll(ctx sdk.Context, pollMeta exported.PollMeta, snapshotCounter int64, expireAt int64) error {
	poll := k.GetPoll(ctx, pollMeta)

	switch {
	case poll != nil && !poll.HasExpired(ctx):
		return fmt.Errorf("poll %s already exists and has not expired yet", pollMeta.String())
	case poll != nil && poll.GetResult() != nil:
		return fmt.Errorf("poll %s has already got result", pollMeta.String())
	default:
		k.DeletePoll(ctx, pollMeta)
		k.setPoll(ctx, types.NewPoll(pollMeta, snapshotCounter, expireAt))

		return nil
	}
}

// DeletePoll deletes the specified poll.
func (k Keeper) DeletePoll(ctx sdk.Context, poll exported.PollMeta) {
	// delete poll
	ctx.KVStore(k.storeKey).Delete([]byte(pollPrefix + poll.String()))

	// delete tallied votes index for poll
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(talliedPrefix+poll.String()))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		ctx.KVStore(k.storeKey).Delete(iter.Key())
	}

	// delete voter index for poll
	iter = sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(addrPrefix+poll.String()))
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		ctx.KVStore(k.storeKey).Delete(iter.Key())
	}
}

// TallyVote tallies votes that have been broadcast. Each validator can only vote once per poll.
func (k Keeper) TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollMeta exported.PollMeta, data exported.VotingData) (*types.Poll, error) {
	poll := k.GetPoll(ctx, pollMeta)
	if poll == nil {
		return nil, fmt.Errorf("poll does not exist or is closed")
	}

	valAddress := k.snapshotter.GetPrincipal(ctx, sender)
	if valAddress == nil {
		return nil, fmt.Errorf("account %v is not registered as a validator proxy", sender.String())
	}

	snap, ok := k.snapshotter.GetSnapshot(ctx, poll.ValidatorSnapshotCounter)
	if !ok {
		return nil, fmt.Errorf("no snapshot found for counter %d", poll.ValidatorSnapshotCounter)
	}

	validator, ok := snap.GetValidator(valAddress)
	if !ok {
		return nil, fmt.Errorf("address %s is not eligible to vote in this poll", valAddress.String())
	}

	if k.getHasVoted(ctx, pollMeta, valAddress) {
		return nil, fmt.Errorf("each validator can only vote once")
	}

	// if the poll is already decided there is no need to keep track of further votes
	if poll.Result != nil || poll.Failed {
		return poll, nil
	}

	k.setHasVoted(ctx, pollMeta, valAddress)
	var talliedVote types.TalliedVote
	// check if others match this vote, create a new unique entry if not, simply add voting power if match is found
	i := k.getTalliedVoteIdx(ctx, pollMeta, data)
	if i == indexNotFound {
		talliedVote = types.NewTalliedVote(validator.ShareCount, data)

		poll.Votes = append(poll.Votes, talliedVote)
		k.setTalliedVoteIdx(ctx, pollMeta, data, len(poll.Votes)-1)
	} else {
		// this assignment copies the value, so we need to write it back into the array
		talliedVote = poll.Votes[i]
		talliedVote.Tally = talliedVote.Tally.AddRaw(validator.ShareCount)
		poll.Votes[i] = talliedVote
	}

	threshold := k.GetVotingThreshold(ctx)
	if threshold.IsMet(talliedVote.Tally, snap.TotalShareCount) {
		k.Logger(ctx).Debug(fmt.Sprintf("threshold of %d/%d has been met for %s: %s/%s",
			threshold.Numerator, threshold.Denominator, pollMeta, talliedVote.Tally.String(), snap.TotalShareCount.String()))

		poll.Result = talliedVote.Data
	}

	_, highestTalliedVote := poll.GetHighestTalliedVote()
	votedShareCount := poll.GetVotedShareCount()

	if !threshold.IsMet(highestTalliedVote.Tally.Add(snap.TotalShareCount).Sub(votedShareCount), snap.TotalShareCount) {
		poll.Failed = true
	}

	k.setPoll(ctx, *poll)

	return poll, nil
}

// GetPoll returns the poll given poll meta
func (k Keeper) GetPoll(ctx sdk.Context, pollMeta exported.PollMeta) *types.Poll {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pollPrefix + pollMeta.String()))
	if bz == nil {
		return nil
	}

	var poll types.Poll
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &poll)

	return &poll
}

func (k Keeper) setPoll(ctx sdk.Context, poll types.Poll) {
	ctx.KVStore(k.storeKey).Set([]byte(pollPrefix+poll.Meta.String()), k.cdc.MustMarshalBinaryLengthPrefixed(&poll))
}

// To adhere to the same one-return-value pattern as the other getters return a marker value if not found
// (returning an int with a pointer reference to be able to return nil instead seems bizarre)
func (k Keeper) getTalliedVoteIdx(ctx sdk.Context, poll exported.PollMeta, data exported.VotingData) int {
	// check if there have been identical votes
	bz := ctx.KVStore(k.storeKey).Get(k.talliedVoteKey(poll, data))
	if bz == nil {
		return indexNotFound
	}

	return int(binary.LittleEndian.Uint64(bz))
}

func (k Keeper) setTalliedVoteIdx(ctx sdk.Context, poll exported.PollMeta, data exported.VotingData, i int) {
	voteKey := k.talliedVoteKey(poll, data)
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(i))

	ctx.KVStore(k.storeKey).Set(voteKey, bz)
}

func (k Keeper) getHasVoted(ctx sdk.Context, poll exported.PollMeta, address sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(addrPrefix + poll.String() + address.String()))
}

func (k Keeper) setHasVoted(ctx sdk.Context, poll exported.PollMeta, address sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(addrPrefix+poll.String()+address.String()), []byte{voted})
}

func (k Keeper) talliedVoteKey(poll exported.PollMeta, data exported.VotingData) []byte {
	return []byte(talliedPrefix + poll.String() + k.hash(data))
}

func (k Keeper) hash(data exported.VotingData) string {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(data)
	h := sha256.Sum256(bz)

	return string(h[:])
}
