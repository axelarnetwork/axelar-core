/*
Package keeper manages second layer voting. It caches votes until they are sent out in a batch and tallies the results.
*/
package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

var (
	thresholdKey = utils.KeyFromStr("votingThreshold")
	pollPrefix   = utils.KeyFromStr("poll")
	votesPrefix  = utils.KeyFromStr("votes")
	voterPrefix  = utils.KeyFromStr("voter")
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
	k.getKVStore(ctx).Set(thresholdKey, &threshold)
}

// GetDefaultVotingThreshold returns the default voting power threshold that must be reached to decide a poll
func (k Keeper) GetDefaultVotingThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.getKVStore(ctx).Get(thresholdKey, &threshold)

	return threshold
}

func (k Keeper) NewPoll(ctx sdk.Context, metadata exported.PollMetadata) exported.Poll {
	return types.NewPollWithLogging(metadata.UpdateBlockHeight(ctx.BlockHeight()), k.newPollStore(ctx), k.Logger(ctx))
}

// GetPoll returns an existing poll to record votes
func (k Keeper) GetPoll(ctx sdk.Context, pollKey exported.PollKey) exported.Poll {
	return types.NewPollWithLogging(k.getPollMetadata(ctx, pollKey), k.newPollStore(ctx), k.Logger(ctx))
}

func (k Keeper) getPollMetadata(ctx sdk.Context, pollKey exported.PollKey) exported.PollMetadata {
	var poll exported.PollMetadata
	if ok := k.getKVStore(ctx).Get(pollPrefix.AppendStr(pollKey.String()), &poll); !ok {
		return exported.PollMetadata{State: exported.NonExistent}
	}

	return poll.UpdateBlockHeight(ctx.BlockHeight())
}

func (k Keeper) getKVStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k Keeper) newPollStore(ctx sdk.Context) *pollStore {
	return &pollStore{
		KVStore:     k.getKVStore(ctx),
		logger:      k.Logger(ctx),
		getSnapshot: func(seqNo int64) (snapshot.Snapshot, bool) { return k.snapshotter.GetSnapshot(ctx, seqNo) },
		getPoll:     func(key exported.PollKey) exported.Poll { return k.GetPoll(ctx, key) },
	}
}

var _ types.Store = &pollStore{}

type pollStore struct {
	votesCached    bool
	snapshotCached bool
	utils.KVStore
	logger      log.Logger
	votes       []types.TalliedVote
	snapshot    snapshot.Snapshot
	getSnapshot func(seqNo int64) (snapshot.Snapshot, bool)
	getPoll     func(key exported.PollKey) exported.Poll
}

func (p *pollStore) SetVote(key exported.PollKey, vote types.TalliedVote) {
	// to keep it simple a single write invalidates the cache
	p.votesCached = false

	p.Set(votesPrefix.AppendStr(key.String()).AppendStr(vote.Hash()), &vote)
}

func (p pollStore) GetVote(key exported.PollKey, hash string) (types.TalliedVote, bool) {
	var vote types.TalliedVote
	ok := p.Get(votesPrefix.AppendStr(key.String()).AppendStr(hash), &vote)
	return vote, ok
}

func (p *pollStore) GetVotes(key exported.PollKey) []types.TalliedVote {
	if !p.votesCached {
		iter := p.Iterator(votesPrefix.AppendStr(key.String()))
		defer utils.CloseLogError(iter, p.logger)

		for ; iter.Valid(); iter.Next() {
			var vote types.TalliedVote
			iter.UnmarshalValue(&vote)
			p.votes = append(p.votes, vote)
		}

		p.votesCached = true
	}

	return p.votes
}

func (p pollStore) SetVoted(key exported.PollKey, voter sdk.ValAddress) {
	p.SetRaw(voterPrefix.AppendStr(key.String()).AppendStr(voter.String()), []byte{})
}

func (p pollStore) HasVoted(key exported.PollKey, voter sdk.ValAddress) bool {
	return p.Has(voterPrefix.AppendStr(key.String()).AppendStr(voter.String()))
}

func (p *pollStore) GetShareCount(snapSeqNo int64, address sdk.ValAddress) (int64, bool) {
	if !p.snapshotCached {
		var ok bool
		p.snapshot, ok = p.getSnapshot(snapSeqNo)
		if !ok {
			panic(fmt.Sprintf("snapshot %d not found", snapSeqNo))
		}
		p.snapshotCached = true
	}
	val, ok := p.snapshot.GetValidator(address)
	if !ok {
		return 0, false
	}
	return val.ShareCount, true
}

func (p *pollStore) GetTotalShareCount(snapSeqNo int64) sdk.Int {
	if !p.snapshotCached {
		var ok bool
		p.snapshot, ok = p.getSnapshot(snapSeqNo)
		if !ok {
			panic(fmt.Sprintf("snapshot %d not found", snapSeqNo))
		}
		p.snapshotCached = true
	}
	return p.snapshot.TotalShareCount
}

func (p pollStore) SetMetadata(metadata exported.PollMetadata) {
	p.Set(pollPrefix.AppendStr(metadata.Key.String()), &metadata)
}

func (p pollStore) GetPoll(key exported.PollKey) exported.Poll {
	return p.getPoll(key)
}

func (p pollStore) DeletePoll(pollKey exported.PollKey) {
	// delete poll metadata
	p.Delete(pollPrefix.AppendStr(pollKey.String()))

	// delete tallied votes index for poll
	iter := p.Iterator(votesPrefix.AppendStr(pollKey.String()))
	defer utils.CloseLogError(iter, p.logger)

	for ; iter.Valid(); iter.Next() {
		p.Delete(iter.GetKey())
	}

	// delete records of past voters
	iter = p.Iterator(voterPrefix.AppendStr(pollKey.String()))
	defer utils.CloseLogError(iter, p.logger)

	for ; iter.Valid(); iter.Next() {
		p.Delete(iter.GetKey())
	}
}
