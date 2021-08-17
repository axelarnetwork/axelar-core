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

// InitializePoll initializes a new poll
func (k Keeper) InitializePoll(ctx sdk.Context, key exported.PollKey, snapshotSeqNo int64, pollProperties ...exported.PollProperty) error {
	metadata := types.NewPollMetaData(key, k.GetDefaultVotingThreshold(ctx), snapshotSeqNo).With(pollProperties...)

	snap, ok := k.snapshotter.GetSnapshot(ctx, metadata.SnapshotSeqNo)
	if !ok {
		return fmt.Errorf("snapshot %d for poll %s must exist", metadata.SnapshotSeqNo, metadata.Key)
	}

	poll := types.NewPoll(metadata, k.newPollStore(ctx, metadata.Key, snap))

	return poll.WithLogging(k.Logger(ctx)).Initialize()
}

// GetPoll returns an existing poll to record votes
func (k Keeper) GetPoll(ctx sdk.Context, pollKey exported.PollKey) exported.Poll {
	metadata, ok := k.getPollMetadata(ctx, pollKey)
	if !ok {
		return &types.Poll{PollMetadata: exported.PollMetadata{State: exported.NonExistent}}
	}

	snap, ok := k.snapshotter.GetSnapshot(ctx, metadata.SnapshotSeqNo)
	if !ok {
		// if the poll already exists the snapshot MUST be there
		panic(fmt.Errorf("could not find snapshot %d for poll %s", metadata.SnapshotSeqNo, pollKey))
	}
	poll := types.NewPoll(metadata, k.newPollStore(ctx, metadata.Key, snap))
	poll.CheckExpiry(ctx.BlockHeight())

	return poll.WithLogging(k.Logger(ctx))
}

func (k Keeper) getPollMetadata(ctx sdk.Context, pollKey exported.PollKey) (exported.PollMetadata, bool) {
	var poll exported.PollMetadata
	if ok := k.getKVStore(ctx).Get(pollPrefix.AppendStr(pollKey.String()), &poll); !ok {
		return exported.PollMetadata{}, false
	}

	return poll, true
}

func (k Keeper) getKVStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k Keeper) newPollStore(ctx sdk.Context, key exported.PollKey, snap snapshot.Snapshot) *pollStore {
	return &pollStore{
		key:      key,
		snapshot: snap,
		KVStore:  k.getKVStore(ctx),
		getPoll:  func(key exported.PollKey) exported.Poll { return k.GetPoll(ctx, key) },
		logger:   k.Logger(ctx),
	}
}

var _ types.Store = &pollStore{}

type pollStore struct {
	votesCached bool
	utils.KVStore
	logger   log.Logger
	votes    []types.TalliedVote
	getPoll  func(key exported.PollKey) exported.Poll
	snapshot snapshot.Snapshot
	key      exported.PollKey
}

func (p *pollStore) GetTotalVoterCount() int64 {
	return int64(len(p.snapshot.Validators))
}

func (p *pollStore) GetTotalShareCount() sdk.Int {
	return p.snapshot.TotalShareCount
}

func (p *pollStore) GetShareCount(voter sdk.ValAddress) (int64, bool) {
	val, ok := p.snapshot.GetValidator(voter)
	if !ok {
		return 0, false
	}
	return val.ShareCount, true
}

func (p *pollStore) SetVote(voter sdk.ValAddress, vote types.TalliedVote) {
	// to keep it simple a single write invalidates the cache
	p.votesCached = false

	p.SetRaw(voterPrefix.AppendStr(p.key.String()).AppendStr(voter.String()), []byte{})
	p.Set(votesPrefix.AppendStr(p.key.String()).AppendStr(vote.Hash()), &vote)
}

func (p pollStore) GetVote(hash string) (types.TalliedVote, bool) {
	var vote types.TalliedVote
	ok := p.Get(votesPrefix.AppendStr(p.key.String()).AppendStr(hash), &vote)
	return vote, ok
}

func (p *pollStore) GetVotes() []types.TalliedVote {
	if !p.votesCached {
		iter := p.Iterator(votesPrefix.AppendStr(p.key.String()))
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

func (p pollStore) HasVoted(voter sdk.ValAddress) bool {
	return p.Has(voterPrefix.AppendStr(p.key.String()).AppendStr(voter.String()))
}

func (p pollStore) SetMetadata(metadata exported.PollMetadata) {
	p.Set(pollPrefix.AppendStr(metadata.Key.String()), &metadata)
}

func (p pollStore) GetPoll(key exported.PollKey) exported.Poll {
	return p.getPoll(key)
}

func (p pollStore) DeletePoll() {
	// delete poll metadata
	p.Delete(pollPrefix.AppendStr(p.key.String()))

	// delete tallied votes index for poll
	iter := p.Iterator(votesPrefix.AppendStr(p.key.String()))
	defer utils.CloseLogError(iter, p.logger)

	for ; iter.Valid(); iter.Next() {
		p.Delete(iter.GetKey())
	}

	// delete records of past voters
	iter = p.Iterator(voterPrefix.AppendStr(p.key.String()))
	defer utils.CloseLogError(iter, p.logger)

	for ; iter.Valid(); iter.Next() {
		p.Delete(iter.GetKey())
	}
}
