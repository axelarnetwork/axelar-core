package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/axelar/types"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	futureVotesKey    = []byte("futureVotesKey")
	publicVotesKey    = []byte("publicVotesKey")
	votingIntervalKey = []byte("votingInterval")
	votingThreshold   = []byte("votingThreshold")
)

type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	broadcaster   types.Broadcaster
	stakingKeeper staking.Keeper
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, stakingKeeper staking.Keeper, client types.Broadcaster) Keeper {
	keeper := Keeper{
		storeKey:      key,
		cdc:           cdc,
		broadcaster:   client,
		stakingKeeper: stakingKeeper,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetFutureVote(ctx sdk.Context, vote exported.FutureVote) {
	k.Logger(ctx).Debug("getting future votes")
	futureVotes := k.getFutureVotes(ctx)

	futureVotes = append(futureVotes, vote)
	k.Logger(ctx).Debug("store future votes")
	k.setFutureVotes(ctx, futureVotes)
}

func (k Keeper) getFutureVotes(ctx sdk.Context) []exported.FutureVote {
	if !ctx.KVStore(k.storeKey).Has(futureVotesKey) {
		return []exported.FutureVote{}
	}
	bz := ctx.KVStore(k.storeKey).Get(futureVotesKey)
	var futureVotes []exported.FutureVote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &futureVotes)
	return futureVotes
}

func (k Keeper) setFutureVotes(ctx sdk.Context, preVoteTxs []exported.FutureVote) {
	ctx.KVStore(k.storeKey).Set(futureVotesKey, k.cdc.MustMarshalBinaryLengthPrefixed(preVoteTxs))
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
	preVotes := k.getFutureVotes(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("unpublished publicVotesKey:%v", len(preVotes)))

	if len(preVotes) == 0 {
		return nil
	}
	var bits []bool
	var votes []types.Vote
	for _, preVote := range preVotes {
		votes = append(votes, types.Vote{
			Tx:            preVote.Tx,
			Confirmations: make(map[string]sdk.ValAddress),
		})
		bits = append(bits, preVote.LocalAccept)
	}

	k.setPublicVotes(ctx, votes)
	// Reset preVotes because this batch is about to be broadcast
	k.setFutureVotes(ctx, []exported.FutureVote{})
	msg := types.NewMsgBatchVote(bits)
	k.Logger(ctx).Debug(fmt.Sprintf("msg:%v", msg))

	return k.broadcaster.Broadcast(ctx, []bcExported.ValidatorMsg{msg})
}

type serializableVote struct {
	Tx           exported.ExternalTx
	ValStrings   []string
	ValAddresses []sdk.ValAddress
}

func (k Keeper) setPublicVotes(ctx sdk.Context, votes []types.Vote) {
	serVotes := mapToSerializable(votes)

	ctx.KVStore(k.storeKey).Set(publicVotesKey, k.cdc.MustMarshalBinaryLengthPrefixed(serVotes))
}

func (k Keeper) getPublicVotes(ctx sdk.Context) []types.Vote {
	if !ctx.KVStore(k.storeKey).Has(publicVotesKey) {
		return nil
	}
	bz := ctx.KVStore(k.storeKey).Get(publicVotesKey)
	var serVotes []serializableVote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &serVotes)

	return mapFromSerializable(serVotes)
}

func mapToSerializable(votes []types.Vote) []serializableVote {
	// The map struct is an unsupported data type for amino, so we need to map it to a list of key-values
	var serVotes []serializableVote
	for _, vote := range votes {
		var valStrings []string
		var valAddresses []sdk.ValAddress

		// WARNING: iterating over a map is not deterministic behaviour,
		// so code down the line must NOT rely on the order of these arrays.
		// Only use the serializableVote struct in the setter and getter.
		for s, a := range vote.Confirmations {
			valStrings = append(valStrings, s)
			valAddresses = append(valAddresses, a)
		}
		serVotes = append(serVotes, serializableVote{
			Tx:           vote.Tx,
			ValStrings:   valStrings,
			ValAddresses: valAddresses,
		})
	}
	return serVotes
}

// see mapToSerializable
func mapFromSerializable(serVotes []serializableVote) []types.Vote {
	var votes []types.Vote
	for _, serVote := range serVotes {
		confirmations := make(map[string]sdk.ValAddress)

		for i := 0; i < len(serVote.ValStrings); i++ {
			confirmations[serVote.ValStrings[i]] = serVote.ValAddresses[i]
		}
		votes = append(votes, types.Vote{
			Tx:            serVote.Tx,
			Confirmations: confirmations,
		})
	}
	return votes
}

// Record all votes from one validator on a batch of transactions
func (k Keeper) RecordVotes(ctx sdk.Context, voter sdk.AccAddress, votes []bool) error {
	validator := k.broadcaster.GetPrincipal(ctx, voter)
	if validator == nil {
		k.Logger(ctx).Error(fmt.Sprintf("connot find voter %v", voter))
		return types.ErrInvalidVoter
	}

	unconfVotes := k.getPublicVotes(ctx)
	if len(votes) != len(unconfVotes) {
		k.Logger(ctx).Debug(fmt.Sprintf("vote length:%v, unconfirmed publicVotesKey: %v", len(votes), len(unconfVotes)))
		return types.ErrInvalidVotes
	}
	for i, vote := range votes {
		k.Logger(ctx).Debug("storing vote confirmation")
		if vote {
			unconfVotes[i].Confirmations[validator.String()] = validator
		}
	}

	k.setPublicVotes(ctx, unconfVotes)
	return nil
}

// Decide if external transactions are accepted based on the number of votes they received
func (k Keeper) TallyCastVotes(ctx sdk.Context) {
	k.Logger(ctx).Debug("decide unconfirmed txs")
	votes := k.getPublicVotes(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("publicVotesKey:%v", len(votes)))

	if len(votes) == 0 {
		return
	}
	totalPower := k.stakingKeeper.GetLastTotalPower(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("total power:%v", totalPower))
	for _, vote := range votes {
		var power = sdk.ZeroInt()
		for _, valAddr := range vote.Confirmations {
			validator := k.stakingKeeper.Validator(ctx, valAddr)
			power = power.AddRaw(validator.GetConsensusPower())
		}
		k.Logger(ctx).Debug(fmt.Sprintf("voting power:%v", power))

		threshold := k.GetVotingThreshold(ctx)
		if threshold.IsMet(power, totalPower) {
			k.confirmTx(ctx, vote.Tx)
		}
	}

	// Transactions have been processed, so reset for the next batch
	k.setPublicVotes(ctx, []types.Vote{})
}

// temporary sanity check and logger until we actually do something with the verified transactions
func (k Keeper) confirmTx(ctx sdk.Context, tx exported.ExternalTx) {
	k.Logger(ctx).Debug(fmt.Sprintf("confirming tx:%v", tx))
	balance := k.getBalance(ctx, tx.Chain)
	balance = balance.Add(tx.Amount)
	k.Logger(ctx).Debug(fmt.Sprintf("balance on %s: %v", tx.Chain, balance.String()))
	k.setBalance(ctx, tx.Chain, balance)
}

func (k Keeper) setBalance(ctx sdk.Context, chain string, balance sdk.DecCoins) {
	ctx.KVStore(k.storeKey).Set([]byte("balance_"+chain), k.cdc.MustMarshalBinaryLengthPrefixed(balance))
}

func (k Keeper) getBalance(ctx sdk.Context, chain string) sdk.DecCoins {
	balanceRaw := ctx.KVStore(k.storeKey).Get([]byte("balance_" + chain))
	if balanceRaw == nil {
		return sdk.NewDecCoins()
	}
	var balance sdk.DecCoins
	k.cdc.MustUnmarshalBinaryLengthPrefixed(balanceRaw, &balance)
	return balance
}

func (k Keeper) GetVotingThreshold(ctx sdk.Context) types.VotingThreshold {
	rawThreshold := ctx.KVStore(k.storeKey).Get(votingThreshold)
	var threshold types.VotingThreshold
	k.cdc.MustUnmarshalBinaryLengthPrefixed(rawThreshold, &threshold)
	return threshold
}

func (k Keeper) SetVotingThreshold(ctx sdk.Context, threshold types.VotingThreshold) {
	ctx.KVStore(k.storeKey).Set(votingThreshold, k.cdc.MustMarshalBinaryLengthPrefixed(threshold))
}
