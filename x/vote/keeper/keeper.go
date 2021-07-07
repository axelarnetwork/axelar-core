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
	// Dummy value: the values do not matter, used as markers
	indexNotFound int = -1
)

var (
	votingThresholdKey = utils.KeyFromStr("votingThreshold")
	pollPrefix         = utils.KeyFromStr("poll")
	talliedPrefix      = utils.KeyFromStr("tallied")
	addrPrefix         = utils.KeyFromStr("addr")

	// Dummy value: the values do not matter, used as markers
	voted = []byte{0}
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
	k.getStore(ctx).Set(votingThresholdKey, &threshold)
}

// GetDefaultVotingThreshold returns the default voting power threshold that must be reached to decide a poll
func (k Keeper) GetDefaultVotingThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.getStore(ctx).Get(votingThresholdKey, &threshold)

	return threshold
}

// InitPoll initializes a new poll. This is the first step of the voting protocol.
// The Keeper only accepts votes for initialized polls.
func (k Keeper) InitPoll(ctx sdk.Context, pollMeta exported.PollMeta, snapshotCounter int64, expireAt int64, threshold ...utils.Threshold) error {
	poll := k.GetPoll(ctx, pollMeta)

	switch {
	case poll != nil && !poll.HasExpired(ctx):
		return fmt.Errorf("poll %s already exists and has not expired yet", pollMeta.String())
	case poll != nil && poll.GetResult() != nil:
		return fmt.Errorf("poll %s has already got result", pollMeta.String())
	default:
		k.DeletePoll(ctx, pollMeta)

		t := k.GetDefaultVotingThreshold(ctx)
		if len(threshold) > 0 {
			t = threshold[0]
		}
		k.setPoll(ctx, types.NewPoll(pollMeta, snapshotCounter, expireAt, t))

		return nil
	}
}

// DeletePoll deletes the specified poll.
func (k Keeper) DeletePoll(ctx sdk.Context, poll exported.PollMeta) {
	// delete poll
	k.getStore(ctx).Delete(pollPrefix.AppendStr(poll.String()))

	// delete tallied votes index for poll
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), talliedPrefix.AppendStr(poll.String()).AsKey())
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		k.getStore(ctx).Delete(utils.KeyFromBz(iter.Key()))
	}

	// delete voter index for poll
	iter = sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), addrPrefix.AppendStr(poll.String()).AsKey())
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		k.getStore(ctx).Delete(utils.KeyFromBz(iter.Key()))
	}
}

// TallyVote tallies votes that have been broadcast. Each validator can only vote once per poll.
func (k Keeper) TallyVote(ctx sdk.Context, sender sdk.AccAddress, pollMeta exported.PollMeta, data codec.ProtoMarshaler) (*types.Poll, error) {
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

	if poll.VotingThreshold.IsMet(talliedVote.Tally, snap.TotalShareCount) {
		k.Logger(ctx).Debug(fmt.Sprintf("threshold of %d/%d has been met for %s: %s/%s",
			poll.VotingThreshold.Numerator, poll.VotingThreshold.Denominator, pollMeta, talliedVote.Tally.String(), snap.TotalShareCount.String()))

		poll.Result = talliedVote.Data
	}

	_, highestTalliedVote := poll.GetHighestTalliedVote()
	votedShareCount := poll.GetVotedShareCount()

	if !poll.VotingThreshold.IsMet(highestTalliedVote.Tally.Add(snap.TotalShareCount).Sub(votedShareCount), snap.TotalShareCount) {
		poll.Failed = true
	}

	k.setPoll(ctx, *poll)

	return poll, nil
}

// GetPoll returns the poll given poll meta
func (k Keeper) GetPoll(ctx sdk.Context, pollMeta exported.PollMeta) *types.Poll {
	var poll types.Poll
	if ok := k.getStore(ctx).Get(pollPrefix.AppendStr(pollMeta.String()), &poll); !ok {
		return nil
	}

	return &poll
}

func (k Keeper) setPoll(ctx sdk.Context, poll types.Poll) {
	k.getStore(ctx).Set(pollPrefix.AppendStr(poll.Meta.String()), &poll)
}

// To adhere to the same one-return-value pattern as the other getters return a marker value if not found
// (returning an int with a pointer reference to be able to return nil instead seems bizarre)
func (k Keeper) getTalliedVoteIdx(ctx sdk.Context, poll exported.PollMeta, data codec.ProtoMarshaler) int {
	// check if there have been identical votes
	key := talliedPrefix.AppendStr(poll.String()).AppendStr(k.hash(data))
	bz := k.getStore(ctx).GetRaw(key)
	if bz == nil {
		return indexNotFound
	}

	return int(binary.LittleEndian.Uint64(bz))
}

func (k Keeper) setTalliedVoteIdx(ctx sdk.Context, poll exported.PollMeta, data codec.ProtoMarshaler, i int) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(i))

	key := talliedPrefix.AppendStr(poll.String()).AppendStr(k.hash(data))
	k.getStore(ctx).SetRaw(key, bz)
}

func (k Keeper) getHasVoted(ctx sdk.Context, poll exported.PollMeta, address sdk.ValAddress) bool {
	return k.getStore(ctx).Has(addrPrefix.AppendStr(poll.String()).AppendStr(address.String()))
}

func (k Keeper) setHasVoted(ctx sdk.Context, poll exported.PollMeta, address sdk.ValAddress) {
	k.getStore(ctx).SetRaw(addrPrefix.AppendStr(poll.String()).AppendStr(address.String()), voted)
}

func (k Keeper) hash(data codec.ProtoMarshaler) string {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(data)
	h := sha256.Sum256(bz)

	return string(h[:])
}

func (k Keeper) getStore(ctx sdk.Context) utils.NormalizedKVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
