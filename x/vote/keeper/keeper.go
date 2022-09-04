/*
Package keeper manages second layer voting. It caches votes until they are sent out in a batch and tallies the results.
*/
package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/proto"
)

var (
	pollPrefix  = key.FromStr("poll")
	votesPrefix = key.FromStr("votes")
	countKey    = key.FromStr("count")

	pollQueueName = "pending_poll_queue"

	// Deprecated
	voterPrefix = key.FromStr("voter")
)

const (
	voteCostPerMaintainer = storetypes.Gas(20000)
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

// InitializePoll creates a poll with the given poll builder
func (k Keeper) InitializePoll(ctx sdk.Context, pollBuilder exported.PollBuilder) (exported.PollID, error) {
	pollMetadata, err := pollBuilder.ID(k.nextPollID(ctx)).Build(ctx.BlockHeight())
	if err != nil {
		return 0, err
	}

	ctx.GasMeter().ConsumeGas(voteCostPerMaintainer*uint64(len(pollMetadata.Snapshot.GetParticipantAddresses())), "initialize poll")

	k.GetPollQueue(ctx).EnqueueNew(pollPrefix.Append(key.From(pollMetadata.ID)), &pollMetadata)

	poll := newPoll(ctx, k, pollMetadata)
	poll.Logger().Info("created poll")

	return pollMetadata.ID, nil
}

// GetPoll returns an existing poll to record votes
func (k Keeper) GetPoll(ctx sdk.Context, id exported.PollID) (exported.Poll, bool) {
	metadata, ok := k.getPollMetadata(ctx, id)
	if !ok {
		return nil, false
	}

	return newPoll(ctx, k, metadata), true
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
	return utils.NewGeneralKVQueue(
		pollQueueName,
		k.getKVStore(ctx),
		k.Logger(ctx),
		func(value codec.ProtoMarshaler) utils.Key {
			metadata := value.(*exported.PollMetadata)
			bz := make([]byte, 8)
			binary.BigEndian.PutUint64(bz, uint64(metadata.ExpiresAt))

			return utils.KeyFromBz(bz)
		},
	)
}

// DeletePoll deletes the poll with the given ID
func (k Keeper) DeletePoll(ctx sdk.Context, pollID exported.PollID) {
	// delete poll metadata
	k.getKVStore(ctx).DeleteNew(pollPrefix.Append(key.From(pollID)))

	// delete tallied votes index for poll
	iter := k.getKVStore(ctx).IteratorNew(votesPrefix.Append(key.From(pollID)))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		k.getKVStore(ctx).DeleteNew(iter.GetNewKey())
	}
}

func (k Keeper) nextPollID(ctx sdk.Context) exported.PollID {
	var val gogoprototypes.UInt64Value
	k.getKVStore(ctx).GetNew(countKey, &val)
	defer k.getKVStore(ctx).SetNew(countKey, &gogoprototypes.UInt64Value{Value: val.Value + 1})

	return exported.PollID(val.Value)
}

func (k Keeper) setPollMetadata(ctx sdk.Context, metadata exported.PollMetadata) {
	k.getKVStore(ctx).SetNew(pollPrefix.Append(key.From(metadata.ID)), &metadata)
}

func (k Keeper) getPollMetadata(ctx sdk.Context, id exported.PollID) (exported.PollMetadata, bool) {
	var poll exported.PollMetadata
	if ok := k.getKVStore(ctx).GetNew(pollPrefix.Append(key.From(id)), &poll); !ok {
		return exported.PollMetadata{}, false
	}

	return poll, true
}

func (k Keeper) getPollMetadatas(ctx sdk.Context) []exported.PollMetadata {
	return utils.GetValues[exported.PollMetadata](k.getKVStore(ctx), pollPrefix)
}

func (k Keeper) getKVStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}

func (k Keeper) getTalliedVotes(ctx sdk.Context, id exported.PollID) []types.TalliedVote {
	return utils.GetValues[types.TalliedVote](k.getKVStore(ctx), votesPrefix.Append(key.From(id)))
}

func (k Keeper) setTalliedVote(ctx sdk.Context, talliedVote types.TalliedVote) {
	k.getKVStore(ctx).SetNew(
		votesPrefix.
			Append(key.From(talliedVote.PollID)).
			Append(key.FromBz(proto.Hash(talliedVote.Data.GetCachedValue().(codec.ProtoMarshaler)))),
		&talliedVote,
	)
}

func (k Keeper) getTalliedVote(ctx sdk.Context, pollID exported.PollID, dataHash []byte) (talliedVote types.TalliedVote, ok bool) {
	return talliedVote, k.getKVStore(ctx).GetNew(
		votesPrefix.
			Append(key.From(pollID)).
			Append(key.FromBz(dataHash)),
		&talliedVote,
	)
}
