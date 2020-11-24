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
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	stExported "github.com/axelarnetwork/axelar-core/x/staking/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

const (
	pendingBallotKey   = "ballot"
	votingIntervalKey  = "votingInterval"
	votingThresholdKey = "votingThreshold"

	// Dummy values: the values do not matter, used as markers
	voted         byte = 0
	indexNotFound      = -1
)

type Keeper struct {
	subjectiveStore store.SubjectiveStore
	storeKey        sdk.StoreKey
	cdc             *codec.Codec
	broadcaster     types.Broadcaster
	staker          stExported.Staker
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, subjectiveStore store.SubjectiveStore, staker stExported.Staker, broadcaster types.Broadcaster) Keeper {
	keeper := Keeper{
		subjectiveStore: subjectiveStore,
		storeKey:        key,
		cdc:             cdc,
		broadcaster:     broadcaster,
		staker:          staker,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetVotingInterval sets the interval in which a ballot is supposed to be broadcast
func (k Keeper) SetVotingInterval(ctx sdk.Context, votingInterval int64) {
	ctx.KVStore(k.storeKey).Set([]byte(votingIntervalKey), k.cdc.MustMarshalBinaryLengthPrefixed(votingInterval))
}

// GetVotingInterval returns the interval in which a ballot is supposed to be broadcast
func (k Keeper) GetVotingInterval(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(votingIntervalKey))

	var interval int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &interval)
	return interval
}

// SetVotingThreshold sets the voting power threshold that must be reached to decide a poll
func (k Keeper) SetVotingThreshold(ctx sdk.Context, threshold types.VotingThreshold) {
	ctx.KVStore(k.storeKey).Set([]byte(votingThresholdKey), k.cdc.MustMarshalBinaryLengthPrefixed(threshold))
}

// GetVotingThreshold returns the voting power threshold that must be reached to decide a poll
func (k Keeper) GetVotingThreshold(ctx sdk.Context) types.VotingThreshold {
	rawThreshold := ctx.KVStore(k.storeKey).Get([]byte(votingThresholdKey))
	var threshold types.VotingThreshold
	k.cdc.MustUnmarshalBinaryLengthPrefixed(rawThreshold, &threshold)
	return threshold
}

// InitPoll initializes a new poll. This is the first step of the voting protocol.
// The Keeper only accepts votes for initialized polls.
func (k Keeper) InitPoll(ctx sdk.Context, poll exported.PollMeta) error {
	if ctx.KVStore(k.storeKey).Has([]byte(poll.String())) {
		return fmt.Errorf("poll with same name already exists")
	}

	ctx.KVStore(k.storeKey).Set([]byte(poll.String()), k.cdc.MustMarshalBinaryLengthPrefixed(types.Poll{Meta: poll}))
	return nil
}

// Vote readies a vote to be broadcast to the entire network.
// Votes are only valid if they correspond to a previously initialized poll.
// Depending on the voting interval multiple votes might be batched together into a ballot before broadcasting.
func (k Keeper) Vote(ctx sdk.Context, vote exported.MsgVote) error {
	if !ctx.KVStore(k.storeKey).Has([]byte(vote.Poll().String())) {
		return fmt.Errorf("no poll registered with the given id")
	}

	ballot := k.getPendingBallot()
	for _, prevVote := range ballot.Votes {
		if prevVote.Poll() == vote.Poll() {
			return fmt.Errorf(fmt.Sprintf("ballot already contains vote for poll %s", vote.Poll()))
		}
	}
	ballot.Votes = append(ballot.Votes, vote)
	k.Logger(ctx).Debug(fmt.Sprintf("new vote for poll %s, data hash: %s",
		vote.Poll().String(), k.hash(vote.Data())))
	k.setPendingBallot(ballot)

	return nil
}

// SendBallot broadcasts all unpublished votes to the entire network.
func (k Keeper) SendBallot(ctx sdk.Context) {
	ballot := k.getPendingBallot()
	k.Logger(ctx).Debug(fmt.Sprintf("unpublished votes:%v", len(ballot.Votes)))

	if len(ballot.Votes) == 0 {
		return
	}

	// Reset ballot for the next round
	k.setPendingBallot(types.MsgBallot{})

	// Broadcast is a subjective action, so it must not cause non-deterministic changes to the tx execution.
	// Because of this and to prevent a deadlock it needs to run in its own goroutine without any callbacks.
	go func(logger log.Logger) {
		err := k.broadcaster.Broadcast(ctx, []bcExported.MsgWithSenderSetter{&ballot})
		if err != nil {
			logger.Error(sdkerrors.Wrap(err, "broadcasting ballot failed").Error())
		} else {
			logger.Debug("broadcasting ballot")
		}
	}(k.Logger(ctx))
}

// TallyVote tallies votes that have been broadcast. Each validator can only vote once per poll.
func (k Keeper) TallyVote(ctx sdk.Context, vote exported.MsgVote) error {
	valAddress := k.broadcaster.GetPrincipal(ctx, vote.GetSigners()[0])
	if valAddress == nil {
		err := fmt.Errorf("account %v is not registered as a validator proxy", vote.GetSigners()[0])
		k.Logger(ctx).Error(err.Error())
		return err
	}
	validator := k.staker.Validator(ctx, valAddress)
	if validator == nil {
		return fmt.Errorf("address does not belong to an account in the validator set")
	}

	poll := k.getPoll(ctx, vote.Poll())
	if poll == nil {
		return fmt.Errorf("poll does not exist or is closed")
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
		talliedVote = poll.Votes[i]
		talliedVote.Tally = talliedVote.Tally.AddRaw(validator.GetConsensusPower())
	}

	threshold := k.GetVotingThreshold(ctx)
	totalPower := k.staker.GetLastTotalPower(ctx)
	if threshold.IsMet(talliedVote.Tally, totalPower) {
		k.Logger(ctx).Debug(fmt.Sprintf("threshold of %d/%d has been met for %s: %s/%s",
			threshold.Numerator, threshold.Denominator, vote.Poll(), talliedVote.Tally.String(), totalPower.String()))
		poll.Result = types.VoteResult{
			PollMeta:   poll.Meta,
			VotingData: talliedVote.Data,
		}
	}

	k.setPoll(ctx, *poll)
	return nil
}

// Result returns the decided outcome of a poll. Returns nil if the poll is still undecided or does not exist.
func (k Keeper) Result(ctx sdk.Context, pollMeta exported.PollMeta) exported.Vote {
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
func (k Keeper) getPendingBallot() types.MsgBallot {
	bz := k.subjectiveStore.Get([]byte(pendingBallotKey))
	if bz == nil {
		return types.MsgBallot{}
	}
	var ballot types.MsgBallot
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &ballot)
	return ballot
}

// See getPendingBallot
func (k Keeper) setPendingBallot(ballot types.MsgBallot) {
	k.subjectiveStore.Set([]byte(pendingBallotKey), k.cdc.MustMarshalBinaryLengthPrefixed(ballot))
}

// using a pointer reference to adhere to the pattern of returning nil if value is not found
func (k Keeper) getPoll(ctx sdk.Context, pollMeta exported.PollMeta) *types.Poll {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pollMeta.String()))
	if bz == nil {
		return nil
	}

	var poll types.Poll
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &poll)
	return &poll
}

func (k Keeper) setPoll(ctx sdk.Context, poll types.Poll) {
	ctx.KVStore(k.storeKey).Set([]byte(poll.Meta.String()), k.cdc.MustMarshalBinaryLengthPrefixed(poll))
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
	return ctx.KVStore(k.storeKey).Has(append([]byte(poll.String()), address...))
}

func (k Keeper) setHasVoted(ctx sdk.Context, poll exported.PollMeta, address sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set(append([]byte(poll.String()), address...), []byte{voted})
}

func (k Keeper) talliedVoteKey(vote exported.MsgVote) []byte {
	return []byte(vote.Poll().String() + k.hash(vote.Data()))
}

func (k Keeper) hash(data exported.VotingData) string {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(data)
	h := sha256.Sum256(bz)
	return string(h[:])
}
