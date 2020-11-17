package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/store"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
	"github.com/axelarnetwork/axelar-core/x/voting/types"
)

const (
	futureVotesKey     = "futureVotes"
	publicVotesKey     = "publicVotes"
	votingIntervalKey  = "votingInterval"
	votingThresholdKey = "votingThreshold"

	// Prefix to partition the key space
	tx_ = "tx_"

	// Dummy values: in some instances the kv store is used as a hash set, so the value does not matter
	voted     byte = 0
	confirmed byte = 0
)

/*
This package manages second layer voting. It caches votes
*/
type Keeper struct {
	subjectiveStore store.SubjectiveStore
	storeKey        sdk.StoreKey
	cdc             *codec.Codec
	broadcaster     types.Broadcaster
	staker          types.Staker
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, subjectiveStore store.SubjectiveStore, staker types.Staker, broadcaster types.Broadcaster) Keeper {
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

// Record all votes on the ballot of a single validator
func (k Keeper) ProcessBallot(ctx sdk.Context, voter sdk.AccAddress, ballot []bool) error {
	k.Logger(ctx).Debug("processing ballot")
	validator := k.broadcaster.GetPrincipal(ctx, voter)
	if validator == nil {
		err := fmt.Errorf("cannot find voter %v", voter)
		k.Logger(ctx).Error(err.Error())
		return err
	}

	openPolls := k.getOpenPolls(ctx)
	if len(ballot) != len(openPolls) {
		k.Logger(ctx).Debug(fmt.Sprintf("votes to record:%v, open polls: %v", len(ballot), len(openPolls)))
		return fmt.Errorf(
			"number of votes on the ballot (%d) did not match number of open polls (%d)",
			len(ballot),
			len(openPolls),
		)
	}

	if k.hasVoted(ctx, validator) {
		return fmt.Errorf("validator %s has already voted", validator.String())
	}

	for i, vote := range ballot {
		k.Logger(ctx).Debug("storing votes")
		openPolls[i].Votes = append(openPolls[i].Votes, types.Vote{
			Validator: validator,
			Confirms:  vote,
		})
	}

	k.setHasVoted(ctx, validator)
	k.setPublicVotes(ctx, openPolls)
	return nil
}

// Decide if external transactions are accepted based on the number of votes they received
func (k Keeper) TallyVotes(ctx sdk.Context) {
	k.Logger(ctx).Debug("tallying votes")
	openPolls := k.getOpenPolls(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("open polls:%v", len(openPolls)))

	if len(openPolls) == 0 {
		return
	}
	totalPower := k.staker.GetLastTotalPower(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("total power:%v", totalPower))
	for _, poll := range openPolls {
		var power = sdk.ZeroInt()
		for _, vote := range poll.Votes {
			validator := k.staker.Validator(ctx, vote.Validator)
			if vote.Confirms {
				power = power.AddRaw(validator.GetConsensusPower())
			}
			// The vote has been tallied, so clear the lookup for the next round
			k.clearHasVoted(ctx, validator.GetOperator())
		}
		k.Logger(ctx).Debug(fmt.Sprintf("voting power for %s: %v", poll.Tx.TxID, power))

		threshold := k.GetVotingThreshold(ctx)
		if threshold.IsMet(power, totalPower) {
			k.confirmTx(ctx, poll.Tx)
		}
	}

	// Transactions have been processed, so reset for the next round
	k.setPublicVotes(ctx, []types.Poll{})
}

func (k Keeper) SetVotingInterval(ctx sdk.Context, votingInterval int64) {
	ctx.KVStore(k.storeKey).Set([]byte(votingIntervalKey), k.cdc.MustMarshalBinaryLengthPrefixed(votingInterval))
}

func (k Keeper) GetVotingInterval(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get([]byte(votingIntervalKey))

	var interval int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &interval)
	k.Logger(ctx).Debug(fmt.Sprintf("voting interval: %v", interval))
	return interval
}

func (k Keeper) SetVotingThreshold(ctx sdk.Context, threshold types.VotingThreshold) {
	ctx.KVStore(k.storeKey).Set([]byte(votingThresholdKey), k.cdc.MustMarshalBinaryLengthPrefixed(threshold))
}

func (k Keeper) GetVotingThreshold(ctx sdk.Context) types.VotingThreshold {
	rawThreshold := ctx.KVStore(k.storeKey).Get([]byte(votingThresholdKey))
	var threshold types.VotingThreshold
	k.cdc.MustUnmarshalBinaryLengthPrefixed(rawThreshold, &threshold)
	return threshold
}
func (k Keeper) setPublicVotes(ctx sdk.Context, votes []types.Poll) {
	ctx.KVStore(k.storeKey).Set([]byte(publicVotesKey), k.cdc.MustMarshalBinaryLengthPrefixed(votes))
}

func (k Keeper) getOpenPolls(ctx sdk.Context) []types.Poll {
	bz := ctx.KVStore(k.storeKey).Get([]byte(publicVotesKey))
	if bz == nil {
		return nil
	}
	var votes []types.Poll
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &votes)
	return votes
}

func (k Keeper) IsVerified(ctx sdk.Context, tx exported.ExternalTx) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(tx_ + string(k.cdc.MustMarshalBinaryLengthPrefixed(tx))))
}

func (k Keeper) confirmTx(ctx sdk.Context, tx exported.ExternalTx) {
	k.Logger(ctx).Debug(fmt.Sprintf("confirming tx: %v", tx))
	key := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(tx_+string(key)), []byte{confirmed})
}

func (k Keeper) hasVoted(ctx sdk.Context, voter sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has(voter.Bytes())
}

func (k Keeper) setHasVoted(ctx sdk.Context, validator sdk.ValAddress) {

	ctx.KVStore(k.storeKey).Set(validator.Bytes(), []byte{voted})
}

func (k Keeper) clearHasVoted(ctx sdk.Context, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Delete(validator.Bytes())
}

/*
//////// Subjective store operations /////////
*/

// SetFutureVote stores the subjective votes of this node
func (k Keeper) SetFutureVote(ctx sdk.Context, vote exported.FutureVote) {
	k.Logger(ctx).Debug("getting future votes")
	futureVotes := k.getFutureVotes()

	futureVotes = append(futureVotes, vote)
	k.Logger(ctx).Debug("store future votes")
	k.setFutureVotes(futureVotes)
}

// BatchVote broadcasts all prerecorded subjective votes to the entire network
func (k Keeper) BatchVote(ctx sdk.Context) error {
	preVotes := k.getFutureVotes()
	k.Logger(ctx).Debug(fmt.Sprintf("unpublished votes:%v", len(preVotes)))

	if len(preVotes) == 0 {
		return nil
	}
	var bits []bool
	var votes []types.Poll
	for _, preVote := range preVotes {
		// prepare vote collecting structure
		votes = append(votes, types.Poll{
			Tx:    preVote.Tx,
			Votes: make([]types.Vote, 0),
		})

		// collect own votes
		bits = append(bits, preVote.LocalAccept)
	}

	k.setPublicVotes(ctx, votes)
	// Reset preVotes because this batch is about to be broadcast
	k.setFutureVotes([]exported.FutureVote{})

	msg := types.NewMsgBatchVote(bits)
	k.Logger(ctx).Debug(fmt.Sprintf("vote: %v", msg))

	// Broadcast is a local action, it must not have any influence on the validity of the message
	go func(logger log.Logger) {
		err := k.broadcaster.Broadcast(ctx, []bcExported.ValidatorMsg{msg})
		if err != nil {
			logger.Error(sdkerrors.Wrap(err, "broadcasting votes failed").Error())
		}
	}(k.Logger(ctx))
	return nil
}

// Because votes may differ between nodes they need to be stored outside the regular kvstore
// (whose hash becomes part of the Merkle tree)
func (k Keeper) getFutureVotes() []exported.FutureVote {
	bz := k.subjectiveStore.Get([]byte(futureVotesKey))
	if bz == nil {
		return []exported.FutureVote{}
	}
	var futureVotes []exported.FutureVote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &futureVotes)
	return futureVotes
}

// See getFutureVotes
func (k Keeper) setFutureVotes(preVoteTxs []exported.FutureVote) {
	k.subjectiveStore.Set([]byte(futureVotesKey), k.cdc.MustMarshalBinaryLengthPrefixed(preVoteTxs))
}
