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

const (
	votingIntervalKey = "votingInterval"
	preVoteTxsKey     = "preVoteTxsKey"
	unconfirmedTxsKey = "unconfirmedTxsKey"
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

	ctx.KVStore(k.storeKey).Set([]byte(address.Address), []byte(address.Chain))

	return nil
}

func (k Keeper) getBridge(chain string) (types.BridgeKeeper, error) {
	br, ok := k.bridges[chain]
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrInvalidChain, "%s is not bridged", chain)
	}
	return br, nil
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
	preVoteTxs := k.getPreVoteTxs(ctx)

	k.Logger(ctx).Debug("letting bridge verify tx")
	preVoteTxs = append(preVoteTxs, types.PreVote{Tx: tx, LocalAccept: br.VerifyTx(ctx, tx)})
	k.Logger(ctx).Debug("store vote")
	ctx.KVStore(k.storeKey).Set([]byte(preVoteTxsKey), k.cdc.MustMarshalBinaryLengthPrefixed(preVoteTxs))
	return nil
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

func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

func (k Keeper) getPreVoteTxs(ctx sdk.Context) []types.PreVote {
	if !ctx.KVStore(k.storeKey).Has([]byte(preVoteTxsKey)) {
		return nil
	}
	bz := ctx.KVStore(k.storeKey).Get([]byte(preVoteTxsKey))
	var preVoteTxs []types.PreVote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &preVoteTxs)
	return preVoteTxs
}

func (k Keeper) getVotes(ctx sdk.Context) []types.Vote {
	if !ctx.KVStore(k.storeKey).Has([]byte(unconfirmedTxsKey)) {
		return nil
	}
	bz := ctx.KVStore(k.storeKey).Get([]byte(unconfirmedTxsKey))
	var votes []types.Vote
	k.cdc.MustUnmarshalJSON(bz, &votes)
	return votes
}

func (k Keeper) BatchVote(ctx sdk.Context) error {
	preVotes := k.getPreVoteTxs(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("unpublished votes:%v", len(preVotes)))

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

	ctx.KVStore(k.storeKey).Set([]byte(unconfirmedTxsKey), k.cdc.MustMarshalJSON(votes))
	ctx.KVStore(k.storeKey).Set([]byte(preVoteTxsKey), k.cdc.MustMarshalBinaryLengthPrefixed([]types.PreVote{}))
	msg := types.NewMsgBatchVote(bits)
	k.Logger(ctx).Debug(fmt.Sprintf("msg:%v", msg))

	return k.broadcaster.Broadcast(ctx, []bcExported.ValidatorMsg{msg})
}

func (k Keeper) RecordVotes(ctx sdk.Context, voter sdk.AccAddress, votes []bool) error {
	if !ctx.KVStore(k.storeKey).Has(voter) {
		k.Logger(ctx).Debug(fmt.Sprintf("connot find voter %v", voter))
		return types.ErrInvalidVoter
	}
	unconfVotes := k.getVotes(ctx)

	if len(votes) != len(unconfVotes) {
		k.Logger(ctx).Debug(fmt.Sprintf("vote length:%v, unconfirmed votes: %v", len(votes), len(unconfVotes)))
		return types.ErrInvalidVotes
	}
	validator := k.broadcaster.GetPrincipal(ctx, voter)
	for i, vote := range votes {
		k.Logger(ctx).Debug("storing vote confirmation")
		if vote {
			if unconfVotes[i].Confirmations == nil {
				unconfVotes[i].Confirmations = make(map[string]sdk.ValAddress)
			}
			unconfVotes[i].Confirmations[validator.String()] = validator
		}
	}

	ctx.KVStore(k.storeKey).Set([]byte(unconfirmedTxsKey), k.cdc.MustMarshalJSON(unconfVotes))
	return nil
}

func (k Keeper) DecideUnconfirmedTxs(ctx sdk.Context) {
	k.Logger(ctx).Debug("decide unconfirmed txs")
	votes := k.getVotes(ctx)
	k.Logger(ctx).Debug(fmt.Sprintf("votes:%v", len(votes)))

	if len(votes) == 0 {
		return
	}
	// TODO: int64 might not be large enough, so this might panic. Check if voting power is bounded
	totalPower := k.stakingKeeper.GetLastTotalPower(ctx).Int64()
	k.Logger(ctx).Debug(fmt.Sprintf("total power:%v", totalPower))
	for _, vote := range votes {
		var power int64 = 0
		for _, valAddr := range vote.Confirmations {
			validator := k.stakingKeeper.Validator(ctx, valAddr)
			power = power + validator.GetConsensusPower()
		}
		k.Logger(ctx).Debug(fmt.Sprintf("voting power:%v", power))
		// tx has 2/3 majority vote to confirm
		if 3*power > 2*totalPower {
			k.confirmTx(ctx, vote.Tx)
		}
	}

	ctx.KVStore(k.storeKey).Set([]byte(unconfirmedTxsKey), k.cdc.MustMarshalJSON([]types.Vote{}))
}

func (k Keeper) confirmTx(ctx sdk.Context, tx exported.ExternalTx) {
	k.Logger(ctx).Debug(fmt.Sprintf("confirming tx:%v", tx))
	balance := k.getBalance(ctx, tx.Chain)
	balance = balance.Add(tx.Amount)
	k.Logger(ctx).Debug(fmt.Sprintf("balance on %s: %v", tx.Chain, balance.String()))
	ctx.KVStore(k.storeKey).Set([]byte("balance_"+tx.Chain), k.cdc.MustMarshalBinaryLengthPrefixed(balance))
}

func (k Keeper) getBalance(ctx sdk.Context, chain string) sdk.Coins {
	balanceRaw := ctx.KVStore(k.storeKey).Get([]byte("balance_" + chain))
	if balanceRaw == nil {
		return sdk.NewCoins()
	}
	var balance sdk.Coins
	k.cdc.MustUnmarshalBinaryLengthPrefixed(balanceRaw, &balance)
	return balance
}
