package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
	"github.com/axelarnetwork/axelar-core/x/axelar/types"
	bcExported "github.com/axelarnetwork/axelar-core/x/broadcast/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	votingIntervalKey = []byte("votingInterval")
	preVotesKey       = []byte("preVotesKey")
	votesKey          = []byte("votesKey")
	votingThreshold   = []byte("votingThreshold")
)

type Keeper struct {
	bridges       map[string]types.BridgeKeeper
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	broadcaster   types.Broadcaster
	stakingKeeper staking.Keeper
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, bridges map[string]types.BridgeKeeper, stakingKeeper staking.Keeper, client types.Broadcaster) Keeper {
	keeper := Keeper{
		bridges:       bridges,
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

func (k Keeper) TrackAddress(ctx sdk.Context, address exported.ExternalChainAddress) error {
	br, err := k.getBridge(address.Chain)
	if err != nil {
		return err
	}

	if err := br.TrackAddress(ctx, address.Address); err != nil {
		return sdkerrors.Wrapf(err, "bridge to %s is unable to track address", address.Chain)
	}

	k.setTrackedAddress(ctx, address)

	return nil
}

func (k Keeper) getBridge(chain string) (types.BridgeKeeper, error) {
	br, ok := k.bridges[chain]
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrInvalidChain, "%s is not bridged", chain)
	}
	return br, nil
}

func (k Keeper) setTrackedAddress(ctx sdk.Context, address exported.ExternalChainAddress) {
	ctx.KVStore(k.storeKey).Set([]byte(address.Address), []byte(address.Chain))
}

func (k Keeper) GetTrackedAddress(ctx sdk.Context, address string) exported.ExternalChainAddress {
	chain := ctx.KVStore(k.storeKey).Get([]byte(address))
	if chain == nil {
		return exported.ExternalChainAddress{}
	}
	return exported.ExternalChainAddress{
		Chain:   string(chain),
		Address: address,
	}
}

func (k Keeper) VerifyTx(ctx sdk.Context, tx exported.ExternalTx) error {
	k.Logger(ctx).Debug("getting bridge")
	br, err := k.getBridge(tx.Chain)
	if err != nil {
		return err
	}

	k.Logger(ctx).Debug("getting prevote txs")
	preVoteTxs := k.getPreVotes(ctx)

	k.Logger(ctx).Debug("letting bridge verify tx")
	preVoteTxs = append(preVoteTxs, types.PreVote{Tx: tx, LocalAccept: br.VerifyTx(ctx, tx)})
	k.Logger(ctx).Debug("store vote")
	k.setPreVotes(ctx, preVoteTxs)
	return nil
}

func (k Keeper) getPreVotes(ctx sdk.Context) []types.PreVote {
	if !ctx.KVStore(k.storeKey).Has(preVotesKey) {
		return []types.PreVote{}
	}
	bz := ctx.KVStore(k.storeKey).Get(preVotesKey)
	var preVotes []types.PreVote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &preVotes)
	return preVotes
}

func (k Keeper) setPreVotes(ctx sdk.Context, preVoteTxs []types.PreVote) {
	ctx.KVStore(k.storeKey).Set(preVotesKey, k.cdc.MustMarshalBinaryLengthPrefixed(preVoteTxs))
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

// Vote on all cached prevote transactions
func (k Keeper) BatchVote(ctx sdk.Context) error {
	preVotes := k.getPreVotes(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("unpublished votesKey:%v", len(preVotes)))

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

	k.setVotes(ctx, votes)
	// Reset preVotes because this batch is about to be broadcast
	k.setPreVotes(ctx, []types.PreVote{})
	msg := types.NewMsgBatchVote(bits)
	k.Logger(ctx).Debug(fmt.Sprintf("msg:%v", msg))

	return k.broadcaster.Broadcast(ctx, []bcExported.ValidatorMsg{msg})
}

type serializableVote struct {
	Tx           exported.ExternalTx
	ValStrings   []string
	ValAddresses []sdk.ValAddress
}

func (k Keeper) setVotes(ctx sdk.Context, votes []types.Vote) {
	serVotes := mapToSerializable(votes)

	ctx.KVStore(k.storeKey).Set(votesKey, k.cdc.MustMarshalBinaryLengthPrefixed(serVotes))
}

func (k Keeper) getVotes(ctx sdk.Context) []types.Vote {
	if !ctx.KVStore(k.storeKey).Has(votesKey) {
		return nil
	}
	bz := ctx.KVStore(k.storeKey).Get(votesKey)
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

	unconfVotes := k.getVotes(ctx)
	if len(votes) != len(unconfVotes) {
		k.Logger(ctx).Debug(fmt.Sprintf("vote length:%v, unconfirmed votesKey: %v", len(votes), len(unconfVotes)))
		return types.ErrInvalidVotes
	}
	for i, vote := range votes {
		k.Logger(ctx).Debug("storing vote confirmation")
		if vote {
			unconfVotes[i].Confirmations[validator.String()] = validator
		}
	}

	k.setVotes(ctx, unconfVotes)
	return nil
}

// Decide if external transactions are accepted based on the number of votes they received
func (k Keeper) TallyCastVotes(ctx sdk.Context) {
	k.Logger(ctx).Debug("decide unconfirmed txs")
	votes := k.getVotes(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("votesKey:%v", len(votes)))

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
		// tx has 2/3 majority vote to confirm
		if threshold.IsMet(power, totalPower) {
			k.confirmTx(ctx, vote.Tx)
		}
	}

	// Transactions have been processed, so reset for the next batch
	k.setVotes(ctx, []types.Vote{})
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
