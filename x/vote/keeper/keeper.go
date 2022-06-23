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
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/slices"
)

var (
	pollPrefix  = utils.KeyFromStr("poll")
	votesPrefix = utils.KeyFromStr("votes")
	voterPrefix = utils.KeyFromStr("voter")

	countKey = utils.KeyFromStr("count")

	pollQueueName = "pending_poll_queue"
)

// Keeper - the vote module's keeper
type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         codec.BinaryCodec
	paramSpace  paramtypes.Subspace
	snapshotter types.Snapshotter
	staking     types.StakingKeeper
	rewarder    types.Rewarder
	voteRouter  types.VoteRouter
}

// NewKeeper - keeper constructor
func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace, snapshotter types.Snapshotter, staking types.StakingKeeper, rewarder types.Rewarder) Keeper {
	keeper := Keeper{
		cdc:         cdc,
		storeKey:    key,
		paramSpace:  paramSpace.WithKeyTable(types.KeyTable()),
		snapshotter: snapshotter,
		staking:     staking,
		rewarder:    rewarder,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the total set of reward parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)

	return params
}

// SetParams sets the total set of reward parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) getPollCount(ctx sdk.Context) uint64 {
	var val gogoprototypes.UInt64Value
	if !k.getKVStore(ctx).Get(countKey, &val) {
		return 0
	}

	return val.Value
}

func (k Keeper) setPollCount(ctx sdk.Context, count uint64) {
	k.getKVStore(ctx).Set(countKey, &gogoprototypes.UInt64Value{Value: count})
}

func (k Keeper) initializePoll(ctx sdk.Context, voters []exported.Voter, pollProperties ...exported.PollProperty) (exported.PollID, error) {
	pollCount := k.getPollCount(ctx)
	pollID := exported.PollID(pollCount)
	metadata := types.NewPollMetaData(
		pollID,
		k.GetParams(ctx).DefaultVotingThreshold,
		slices.Filter(voters, func(v exported.Voter) bool {
			return v.VotingPower > 0
		})).
		With(pollProperties...)

	poll := types.NewPoll(metadata, k.newPollStore(ctx, metadata.ID)).WithLogger(k.Logger(ctx))
	if err := poll.Initialize(ctx.BlockHeight()); err != nil {
		return 0, err
	}

	k.setPollCount(ctx, pollCount+1)

	k.Logger(ctx).Info("created poll",
		"poll", pollID.String(),
	)

	return pollID, nil
}

// InitializePoll initializes a new poll with the given validators
func (k Keeper) InitializePoll(ctx sdk.Context, voterAddresses []sdk.ValAddress, pollProperties ...exported.PollProperty) (exported.PollID, error) {
	voters := make([]exported.Voter, 0)

	for _, voterAddress := range voterAddresses {
		validator := k.staking.Validator(ctx, voterAddress)
		if validator == nil {
			k.Logger(ctx).Debug(fmt.Sprintf("voter %s is not a validator", voterAddress.String()))
			continue
		}

		voters = append(voters, exported.Voter{Validator: voterAddress, VotingPower: validator.GetConsensusPower(k.staking.PowerReduction(ctx))})
	}

	return k.initializePoll(ctx, voters, pollProperties...)
}

// InitializePollWithSnapshot initializes a new poll with the given snapshot sequence number
func (k Keeper) InitializePollWithSnapshot(ctx sdk.Context, snapshotSeqNo int64, pollProperties ...exported.PollProperty) (exported.PollID, error) {
	snap, ok := k.snapshotter.GetSnapshot(ctx, snapshotSeqNo)
	if !ok {
		return 0, fmt.Errorf("snapshot %d does not exist", snapshotSeqNo)
	}

	voters := make([]exported.Voter, 0)
	for _, validator := range snap.Validators {
		voters = append(voters, exported.Voter{Validator: validator.GetSDKValidator().GetOperator(), VotingPower: validator.ShareCount})
	}

	return k.initializePoll(ctx, voters, pollProperties...)
}

// GetPoll returns an existing poll to record votes
func (k Keeper) GetPoll(ctx sdk.Context, id exported.PollID) exported.Poll {
	metadata, ok := k.getPollMetadata(ctx, id)
	if !ok {
		return &types.Poll{PollMetadata: exported.PollMetadata{State: exported.NonExistent}}
	}

	poll := types.NewPoll(metadata, k.newPollStore(ctx, metadata.ID)).WithLogger(k.Logger(ctx))

	return poll
}

func (k Keeper) getPollMetadata(ctx sdk.Context, id exported.PollID) (exported.PollMetadata, bool) {
	var poll exported.PollMetadata
	if ok := k.getKVStore(ctx).Get(pollPrefix.AppendStr(id.String()), &poll); !ok {
		return exported.PollMetadata{}, false
	}

	return poll, true
}

func (k Keeper) getNonPendingPollMetadatas(ctx sdk.Context) []exported.PollMetadata {
	return slices.Filter(
		k.getPollMetadatas(ctx),
		func(metadata exported.PollMetadata) bool { return !metadata.Is(exported.Pending) },
	)
}

func (k Keeper) getPollMetadatas(ctx sdk.Context) []exported.PollMetadata {
	var pollMetadatas []exported.PollMetadata

	iter := k.getKVStore(ctx).Iterator(pollPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var pollMetadata exported.PollMetadata
		iter.UnmarshalValue(&pollMetadata)
		pollMetadatas = append(pollMetadatas, pollMetadata)
	}

	return pollMetadatas
}

func (k Keeper) getKVStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k Keeper) newPollStore(ctx sdk.Context, id exported.PollID) *pollStore {
	return &pollStore{
		id:      id,
		KVStore: k.getKVStore(ctx),
		getPoll: func(id exported.PollID) exported.Poll { return k.GetPoll(ctx, id) },
		logger:  k.Logger(ctx),
	}
}

// SetVoteRouter sets the vote router. It will panic if called more than once
func (k *Keeper) SetVoteRouter(router types.VoteRouter) {
	if k.voteRouter != nil {
		panic("router already set")
	}

	k.voteRouter = router

	// In order to avoid invalid or non-deterministic behavior, we seal the router immediately
	// to prevent additional handlers from being registered after the keeper is initialized.
	k.voteRouter.Seal()
}

// GetVoteRouter returns the nexus router. If no router was set, it returns a (sealed) router with no handlers
func (k Keeper) GetVoteRouter() types.VoteRouter {
	if k.voteRouter == nil {
		k.SetVoteRouter(types.NewRouter())
	}

	return k.voteRouter
}

// GetPollQueue returns the poll queue
func (k Keeper) GetPollQueue(ctx sdk.Context) utils.KVQueue {
	return getPollQueue(k.getKVStore(ctx), k.Logger(ctx))
}

func getPollQueue(store utils.KVStore, logger log.Logger) utils.KVQueue {
	return utils.NewGeneralKVQueue(
		pollQueueName,
		store,
		logger,
		func(value codec.ProtoMarshaler) utils.Key {
			metadata := value.(*exported.PollMetadata)
			bz := make([]byte, 8)
			binary.BigEndian.PutUint64(bz, uint64(metadata.ExpiresAt))

			return utils.KeyFromBz(bz)
		},
	)
}

var _ types.Store = &pollStore{}

type pollStore struct {
	votesCached bool
	utils.KVStore
	logger  log.Logger
	votes   []types.TalliedVote
	getPoll func(key exported.PollID) exported.Poll
	id      exported.PollID
}

func hash(data codec.ProtoMarshaler) string {
	bz, err := data.Marshal()
	if err != nil {
		panic(err)
	}
	h := sha256.Sum256(bz)

	return string(h[:])
}

func (p *pollStore) SetVote(voter sdk.ValAddress, data codec.ProtoMarshaler, votingPower int64, isLate bool) {
	dataHash := hash(data)

	var talliedVote types.TalliedVote
	if existingVote, ok := p.GetVote(dataHash); !ok {
		talliedVote = types.NewTalliedVote(voter, votingPower, data)
	} else {
		talliedVote = existingVote
		talliedVote.Tally = talliedVote.Tally.AddRaw(votingPower)
		talliedVote.Voters = append(talliedVote.Voters, voter)
	}

	// to keep it simple a single write invalidates the cache
	p.votesCached = false

	p.Set(voterPrefix.AppendStr(p.id.String()).AppendStr(voter.String()), &types.VoteRecord{Voter: voter, IsLate: isLate})
	p.Set(votesPrefix.AppendStr(p.id.String()).AppendStr(dataHash), &talliedVote)
}

func (p pollStore) GetVote(hash string) (types.TalliedVote, bool) {
	var vote types.TalliedVote
	ok := p.Get(votesPrefix.AppendStr(p.id.String()).AppendStr(hash), &vote)
	return vote, ok
}

func (p *pollStore) GetVotes() []types.TalliedVote {
	if !p.votesCached {
		p.votes = []types.TalliedVote{}

		iter := p.Iterator(votesPrefix.AppendStr(p.id.String()))
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
	return p.Has(voterPrefix.AppendStr(p.id.String()).AppendStr(voter.String()))
}

func (p pollStore) HasVotedLate(voter sdk.ValAddress) bool {
	var voteRecord types.VoteRecord

	return p.Get(voterPrefix.AppendStr(p.id.String()).AppendStr(voter.String()), &voteRecord) && voteRecord.IsLate
}

func getPollMetadataKey(metadata exported.PollMetadata) utils.Key {
	return pollPrefix.AppendStr(metadata.ID.String())
}

func (p pollStore) SetMetadata(metadata exported.PollMetadata) {
	p.Set(getPollMetadataKey(metadata), &metadata)
}

func (p pollStore) EnqueuePoll(metadata exported.PollMetadata) {
	getPollQueue(p.KVStore, p.logger).Enqueue(getPollMetadataKey(metadata), &metadata)
}

func (p pollStore) GetPoll(id exported.PollID) exported.Poll {
	return p.getPoll(id)
}

func (p pollStore) DeletePoll() {
	// delete poll metadata
	p.Delete(pollPrefix.AppendStr(p.id.String()))

	// delete tallied votes index for poll
	iter := p.Iterator(votesPrefix.AppendStr(p.id.String()))
	defer utils.CloseLogError(iter, p.logger)

	for ; iter.Valid(); iter.Next() {
		p.Delete(iter.GetKey())
	}

	// delete records of past voters
	iter = p.Iterator(voterPrefix.AppendStr(p.id.String()))
	defer utils.CloseLogError(iter, p.logger)

	for ; iter.Valid(); iter.Next() {
		p.Delete(iter.GetKey())
	}
}
