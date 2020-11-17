package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/store"
	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/axelar/types"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	futureVotesKey     = []byte("futureVotesKey")
	publicVotesKey     = []byte("publicVotesKey")
	votingIntervalKey  = []byte("votingIntervalKey")
	votingThresholdKey = []byte("votingThresholdKey")
	txKey              = []byte("txKey")
)

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

// SetFutureVote stores the subjective votes of this node
func (k Keeper) SetFutureVote(ctx sdk.Context, vote exported.FutureVote) {
	k.Logger(ctx).Debug("getting future votes")
	futureVotes := k.getFutureVotes()

	futureVotes = append(futureVotes, vote)
	k.Logger(ctx).Debug("store future votes")
	k.setFutureVotes(futureVotes)
}

// Because votes may differ between nodes they need to be stored outside the regular kvstore
// (whose hash becomes part of the Merkle tree)
func (k Keeper) getFutureVotes() []exported.FutureVote {
	bz := k.subjectiveStore.Get(futureVotesKey)
	if bz == nil {
		return []exported.FutureVote{}
	}
	var futureVotes []exported.FutureVote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &futureVotes)
	return futureVotes
}

// See getFutureVotes
func (k Keeper) setFutureVotes(preVoteTxs []exported.FutureVote) {
	k.subjectiveStore.Set(futureVotesKey, k.cdc.MustMarshalBinaryLengthPrefixed(preVoteTxs))
}

func (k Keeper) SetVotingInterval(ctx sdk.Context, votingInterval int64) {
	ctx.KVStore(k.storeKey).Set(votingIntervalKey, k.cdc.MustMarshalBinaryLengthPrefixed(votingInterval))
}

func (k Keeper) GetVotingInterval(ctx sdk.Context) int64 {
	bz := ctx.KVStore(k.storeKey).Get(votingIntervalKey)

	var interval int64
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &interval)
	k.Logger(ctx).Debug(fmt.Sprintf("voting interval: %v", interval))
	return interval
}

// Broadcast the batched future votes
func (k Keeper) BatchVote(ctx sdk.Context) error {
	preVotes := k.getFutureVotes()
	k.Logger(ctx).Debug(fmt.Sprintf("unpublished votes:%v", len(preVotes)))

	if len(preVotes) == 0 {
		return nil
	}
	var bits []bool
	var votes []types.Vote
	for _, preVote := range preVotes {
		// prepare vote collecting structure
		votes = append(votes, types.Vote{
			Tx:            preVote.Tx,
			Confirmations: make([]types.Confirmation, 0),
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

func (k Keeper) setPublicVotes(ctx sdk.Context, votes []types.Vote) {
	ctx.KVStore(k.storeKey).Set(publicVotesKey, k.cdc.MustMarshalBinaryLengthPrefixed(votes))
}

func (k Keeper) getPublicVotes(ctx sdk.Context) []types.Vote {
	bz := ctx.KVStore(k.storeKey).Get(publicVotesKey)
	if bz == nil {
		return nil
	}
	var votes []types.Vote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &votes)
	return votes
}

// Record all votes from one validator on a batch of transactions
func (k Keeper) RecordVotes(ctx sdk.Context, voter sdk.AccAddress, votes []bool) error {
	k.Logger(ctx).Debug("recording vote")
	validator := k.broadcaster.GetPrincipal(ctx, voter)
	if validator == nil {
		k.Logger(ctx).Error(fmt.Sprintf("cannot find voter %v", voter))
		return types.ErrInvalidVoter
	}

	unconfirmedVotes := k.getPublicVotes(ctx)
	if len(votes) != len(unconfirmedVotes) {
		k.Logger(ctx).Debug(fmt.Sprintf("votes to record:%v, unconfirmed transactions: %v", len(votes), len(unconfirmedVotes)))
		return types.ErrInvalidVotes
	}
	for i, vote := range votes {
		k.Logger(ctx).Debug("storing vote confirmation")
		// ignore votes from validators that already voted this round
		if !k.hasVoted(ctx, validator) {
			unconfirmedVotes[i].Confirmations = append(unconfirmedVotes[i].Confirmations, types.Confirmation{
				Validator: validator,
				Confirms:  vote,
			})
		}
	}

	k.setHasVoted(ctx, validator)
	k.setPublicVotes(ctx, unconfirmedVotes)
	return nil
}

// Decide if external transactions are accepted based on the number of votes they received
func (k Keeper) TallyCastVotes(ctx sdk.Context) {
	k.Logger(ctx).Debug("tally votes")
	votes := k.getPublicVotes(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("cast votes:%v", len(votes)))

	if len(votes) == 0 {
		return
	}
	totalPower := k.staker.GetLastTotalPower(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("total power:%v", totalPower))
	for _, vote := range votes {
		var power = sdk.ZeroInt()
		for _, confirmation := range vote.Confirmations {
			validator := k.staker.Validator(ctx, confirmation.Validator)
			if confirmation.Confirms {
				power = power.AddRaw(validator.GetConsensusPower())
			}
			// The vote has been tallied, so clear the lookup for the next round
			k.clearHasVoted(ctx, validator.GetOperator())
		}
		k.Logger(ctx).Debug(fmt.Sprintf("voting power for %s: %v", vote.Tx.TxID, power))

		threshold := k.GetVotingThreshold(ctx)
		if threshold.IsMet(power, totalPower) {
			k.confirmTx(ctx, vote.Tx)
		}
	}

	// Transactions have been processed, so reset for the next round
	k.setPublicVotes(ctx, []types.Vote{})
}

func (k Keeper) IsVerified(ctx sdk.Context, tx exported.ExternalTx) bool {
	return ctx.KVStore(k.storeKey).Has(append(txKey, k.cdc.MustMarshalBinaryLengthPrefixed(tx)...))
}

func (k Keeper) confirmTx(ctx sdk.Context, tx exported.ExternalTx) {
	k.Logger(ctx).Debug(fmt.Sprintf("confirming tx: %v", tx))
	key := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set(append(txKey, key...), []byte("confirmed"))
}

func (k Keeper) GetVotingThreshold(ctx sdk.Context) types.VotingThreshold {
	rawThreshold := ctx.KVStore(k.storeKey).Get(votingThresholdKey)
	var threshold types.VotingThreshold
	k.cdc.MustUnmarshalBinaryLengthPrefixed(rawThreshold, &threshold)
	return threshold
}

func (k Keeper) SetVotingThreshold(ctx sdk.Context, threshold types.VotingThreshold) {
	ctx.KVStore(k.storeKey).Set(votingThresholdKey, k.cdc.MustMarshalBinaryLengthPrefixed(threshold))
}

func (k Keeper) hasVoted(ctx sdk.Context, voter sdk.ValAddress) bool {
	return ctx.KVStore(k.storeKey).Has(voter.Bytes())
}

func (k Keeper) setHasVoted(ctx sdk.Context, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Set(validator.Bytes(), []byte("voted"))
}

func (k Keeper) clearHasVoted(ctx sdk.Context, validator sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Delete(validator.Bytes())
}
