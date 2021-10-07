package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	pendingOutpointPrefix    = utils.KeyFromStr("pend_")
	confirmedOutPointPrefix  = utils.KeyFromStr("conf_")
	spentOutPointPrefix      = utils.KeyFromStr("spent_")
	addrPrefix               = utils.KeyFromStr("addr_")
	dustAmtPrefix            = utils.KeyFromStr("dust_")
	signedTxPrefix           = utils.KeyFromStr("signed_tx_")
	unsignedTxPrefix         = utils.KeyFromStr("unsigned_tx_")
	latestSignedTxHashPrefix = utils.KeyFromStr("latest_signed_tx_hash_")
	unconfirmedAmountPrefix  = utils.KeyFromStr("unconfirmed_amount_")

	externalKeyIDsKey        = utils.KeyFromStr("external_key_ids")
	anyoneCanSpendAddressKey = utils.KeyFromStr("anyone_can_spend_address")

	confirmedOutpointQueueName = "confirmed_outpoint"
)

var _ types.BTCKeeper = Keeper{}

// Keeper provides access to all state changes regarding the Bitcoin module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
	params   params.Subspace
}

// NewKeeper returns a new keeper object
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// SetParams sets the bitcoin module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
	anyoneCanSpendAddress := types.NewAnyoneCanSpendAddress(p.Network)
	k.getStore(ctx).Set(anyoneCanSpendAddressKey, &anyoneCanSpendAddress)
}

// GetParams gets the bitcoin module's parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.params.GetParamSet(ctx, &p)
	return p
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAnyoneCanSpendAddress retrieves the anyone-can-spend address
func (k Keeper) GetAnyoneCanSpendAddress(ctx sdk.Context) types.AddressInfo {
	var address types.AddressInfo
	ok := k.getStore(ctx).Get(anyoneCanSpendAddressKey, &address)
	if !ok {
		panic("bitcoin's anyone-can-pay-address isn't set")
	}

	return address
}

// GetRequiredConfirmationHeight returns the minimum number of confirmations a transaction must have on Bitcoin
// before axelar will accept it as confirmed.
func (k Keeper) GetRequiredConfirmationHeight(ctx sdk.Context) uint64 {
	var h uint64
	k.params.Get(ctx, types.KeyConfirmationHeight, &h)
	return h
}

// GetRevoteLockingPeriod returns the lock period for revoting
func (k Keeper) GetRevoteLockingPeriod(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeyRevoteLockingPeriod, &result)

	return result
}

// GetSigCheckInterval returns the block interval after which to check for completed signatures
func (k Keeper) GetSigCheckInterval(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeySigCheckInterval, &result)

	return result
}

// GetNetwork returns the connected Bitcoin network (main, test, regtest)
func (k Keeper) GetNetwork(ctx sdk.Context) types.Network {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return network
}

// GetMinOutputAmount returns the minimum withdrawal threshold
func (k Keeper) GetMinOutputAmount(ctx sdk.Context) btcutil.Amount {
	var coin sdk.DecCoin
	k.params.Get(ctx, types.KeyMinOutputAmount, &coin)

	satoshi, err := types.ToSatoshiCoin(coin)
	if err != nil {
		panic(err)
	}

	return btcutil.Amount(satoshi.Amount.Int64())
}

// GetMaxInputCount returns the max input count
func (k Keeper) GetMaxInputCount(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeyMaxInputCount, &result)

	return result
}

// GetVotingThreshold returns voting threshold
func (k Keeper) GetVotingThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyVotingThreshold, &threshold)

	return threshold
}

// GetMinVoterCount returns minimum voter count for voting
func (k Keeper) GetMinVoterCount(ctx sdk.Context) int64 {
	var minVoterCount int64
	k.params.Get(ctx, types.KeyMinVoterCount, &minVoterCount)

	return minVoterCount
}

// GetMaxSecondaryOutputAmount returns the max secondary output amount
func (k Keeper) GetMaxSecondaryOutputAmount(ctx sdk.Context) btcutil.Amount {
	var coin sdk.DecCoin
	k.params.Get(ctx, types.KeyMaxSecondaryOutputAmount, &coin)

	satoshi, err := types.ToSatoshiCoin(coin)
	if err != nil {
		panic(err)
	}

	return btcutil.Amount(satoshi.Amount.Int64())
}

// GetMasterKeyRetentionPeriod returns the prev master key cycle
func (k Keeper) GetMasterKeyRetentionPeriod(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeyMasterKeyRetentionPeriod, &result)

	return result
}

// GetMasterAddressInternalKeyLockDuration returns the master address lock duration for internal key(s) only to spend
func (k Keeper) GetMasterAddressInternalKeyLockDuration(ctx sdk.Context) time.Duration {
	var result time.Duration
	k.params.Get(ctx, types.KeyMasterAddressInternalKeyLockDuration, &result)

	return result
}

// GetMasterAddressExternalKeyLockDuration returns the master address lock duration for external key(s) only to spend
func (k Keeper) GetMasterAddressExternalKeyLockDuration(ctx sdk.Context) time.Duration {
	var result time.Duration
	k.params.Get(ctx, types.KeyMasterAddressExternalKeyLockDuration, &result)

	return result
}

// GetMaxTxSize returns the max tx size allowed
func (k Keeper) GetMaxTxSize(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeyMaxTxSize, &result)

	return result
}

// SetAddress stores the given address information
func (k Keeper) SetAddress(ctx sdk.Context, address types.AddressInfo) {
	k.getStore(ctx).Set(addrPrefix.Append(utils.LowerCaseKey(address.Address)), &address)
}

// GetAddress returns the address information for the given encoded address
func (k Keeper) GetAddress(ctx sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
	var address types.AddressInfo
	ok := k.getStore(ctx).Get(addrPrefix.Append(utils.LowerCaseKey(encodedAddress)), &address)
	return address, ok
}

// DeleteOutpointInfo deletes a the given outpoint if known
func (k Keeper) DeleteOutpointInfo(ctx sdk.Context, outPoint wire.OutPoint) {
	// delete is a noop if key does not exist
	key := utils.LowerCaseKey(outPoint.String())
	k.getStore(ctx).Delete(confirmedOutPointPrefix.Append(key))
	k.getStore(ctx).Delete(spentOutPointPrefix.Append(key))
}

// GetPendingOutPointInfo returns outpoint information associated with the given poll
func (k Keeper) GetPendingOutPointInfo(ctx sdk.Context, key vote.PollKey) (types.OutPointInfo, bool) {
	var info types.OutPointInfo
	ok := k.getStore(ctx).Get(pendingOutpointPrefix.Append(utils.LowerCaseKey(key.String())), &info)
	return info, ok
}

// GetOutPointInfo returns additional information for the given outpoint
func (k Keeper) GetOutPointInfo(ctx sdk.Context, outPoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
	var info types.OutPointInfo

	key := utils.LowerCaseKey(outPoint.String())
	ok := k.getStore(ctx).Get(confirmedOutPointPrefix.Append(key), &info)
	if ok {
		return info, types.OutPointState_Confirmed, true
	}

	ok = k.getStore(ctx).Get(spentOutPointPrefix.Append(key), &info)
	if ok {
		return info, types.OutPointState_Spent, true
	}

	return types.OutPointInfo{}, 0, false
}

// SetPendingOutpointInfo stores an unconfirmed outpoint.
// Since the information is not yet confirmed the outpoint info is not necessarily unique.
// Therefore we need to store by the poll that confirms/rejects it
func (k Keeper) SetPendingOutpointInfo(ctx sdk.Context, key vote.PollKey, info types.OutPointInfo) {
	k.getStore(ctx).Set(pendingOutpointPrefix.Append(utils.LowerCaseKey(key.String())), &info)
}

// DeletePendingOutPointInfo deletes the outpoint information associated with the given poll
func (k Keeper) DeletePendingOutPointInfo(ctx sdk.Context, key vote.PollKey) {
	k.getStore(ctx).Delete(pendingOutpointPrefix.Append(utils.LowerCaseKey(key.String())))
}

// SetSpentOutpointInfo stores the given outpoint info as spent
func (k Keeper) SetSpentOutpointInfo(ctx sdk.Context, info types.OutPointInfo) {
	key := utils.LowerCaseKey(info.OutPoint)

	k.getStore(ctx).Set(spentOutPointPrefix.Append(key), &info)
}

// SetConfirmedOutpointInfo stores the given outpoint info as confirmed and push it into the queue of given keyID
func (k Keeper) SetConfirmedOutpointInfo(ctx sdk.Context, keyID tss.KeyID, info types.OutPointInfo) {
	key := utils.LowerCaseKey(info.OutPoint)

	k.GetConfirmedOutpointInfoQueueForKey(ctx, keyID).Enqueue(confirmedOutPointPrefix.Append(key), &info)
}

// GetConfirmedOutpointInfoQueueForKey retrieves the outpoint info queue for the given keyID
func (k Keeper) GetConfirmedOutpointInfoQueueForKey(ctx sdk.Context, keyID tss.KeyID) utils.KVQueue {
	queueName := fmt.Sprintf("%s_%s", confirmedOutpointQueueName, keyID)

	return utils.NewBlockHeightKVQueue(queueName, k.getStore(ctx), ctx.BlockHeight(), k.Logger(ctx))
}

// SetUnsignedTx stores an unsigned transaction
func (k Keeper) SetUnsignedTx(ctx sdk.Context, tx types.UnsignedTx) {
	k.getStore(ctx).Set(unsignedTxPrefix.AppendStr(tx.Type.SimpleString()), &tx)
}

// GetUnsignedTx returns the unsigned transaction for the given tx type
func (k Keeper) GetUnsignedTx(ctx sdk.Context, txType types.TxType) (types.UnsignedTx, bool) {
	var result types.UnsignedTx
	if ok := k.getStore(ctx).Get(unsignedTxPrefix.AppendStr(txType.SimpleString()), &result); !ok {
		return types.UnsignedTx{}, false
	}

	return result, true
}

// DeleteUnsignedTx deletes the unsigned transaction for the given tx type
func (k Keeper) DeleteUnsignedTx(ctx sdk.Context, txType types.TxType) {
	k.getStore(ctx).Delete(unsignedTxPrefix.AppendStr(txType.SimpleString()))
}

// SetSignedTx stores the signed transaction for outpoint consolidation
func (k Keeper) SetSignedTx(ctx sdk.Context, tx types.SignedTx) {
	prevSignedTxHash, ok := k.GetLatestSignedTxHash(ctx, tx.Type)
	if ok {
		tx.PrevSignedTxHash = prevSignedTxHash[:]
	} else {
		tx.PrevSignedTxHash = nil
	}

	k.getStore(ctx).Set(signedTxPrefix.Append(utils.LowerCaseKey(tx.GetTx().TxHash().String())), &tx)
}

// GetSignedTx returns the signed transaction for outpoint consolidation
func (k Keeper) GetSignedTx(ctx sdk.Context, txHash chainhash.Hash) (types.SignedTx, bool) {
	var result types.SignedTx
	if ok := k.getStore(ctx).Get(signedTxPrefix.Append(utils.LowerCaseKey(txHash.String())), &result); !ok {
		return types.SignedTx{}, false
	}

	return result, true
}

// SetLatestSignedTxHash stores the tx hash of the most recent transaction signed of the given tx type
func (k Keeper) SetLatestSignedTxHash(ctx sdk.Context, txType types.TxType, txHash chainhash.Hash) {
	k.getStore(ctx).SetRaw(latestSignedTxHashPrefix.AppendStr(txType.SimpleString()), txHash[:])
}

// GetLatestSignedTxHash retrieves the tx hash of the most recent transaction signed of the given tx type
func (k Keeper) GetLatestSignedTxHash(ctx sdk.Context, txType types.TxType) (*chainhash.Hash, bool) {
	bz := k.getStore(ctx).GetRaw(latestSignedTxHashPrefix.AppendStr(txType.SimpleString()))
	if bz == nil {
		return nil, false
	}

	txHash, err := chainhash.NewHash(bz)
	if err != nil {
		panic(err)
	}

	return txHash, true
}

// SetDustAmount stores the dust amount for a destination bitcoin address
func (k Keeper) SetDustAmount(ctx sdk.Context, encodedAddress string, amount btcutil.Amount) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(amount))

	k.getStore(ctx).SetRaw(dustAmtPrefix.Append(utils.LowerCaseKey(encodedAddress)), bz)
}

// GetDustAmount returns the dust amount for a destination bitcoin address
func (k Keeper) GetDustAmount(ctx sdk.Context, encodedAddress string) btcutil.Amount {
	bz := k.getStore(ctx).GetRaw(dustAmtPrefix.Append(utils.LowerCaseKey(encodedAddress)))
	if bz == nil {
		return 0
	}

	return btcutil.Amount(int64(binary.LittleEndian.Uint64(bz)))
}

// DeleteDustAmount deletes the dust amount for a destination bitcoin address
func (k Keeper) DeleteDustAmount(ctx sdk.Context, encodedAddress string) {
	k.getStore(ctx).Delete(dustAmtPrefix.Append(utils.LowerCaseKey(encodedAddress)))
}

// SetUnconfirmedAmount stores the unconfirmed amount for the given key ID
func (k Keeper) SetUnconfirmedAmount(ctx sdk.Context, keyID tss.KeyID, amount btcutil.Amount) {
	if amount < 0 {
		amount = 0
	}

	k.getStore(ctx).Set(unconfirmedAmountPrefix.AppendStr(string(keyID)), &gogoprototypes.Int64Value{Value: int64(amount)})
}

// GetUnconfirmedAmount retrieves the unconfirmed amount for the given key ID
func (k Keeper) GetUnconfirmedAmount(ctx sdk.Context, keyID tss.KeyID) btcutil.Amount {
	var result gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(unconfirmedAmountPrefix.AppendStr(string(keyID)), &result); !ok {
		return 0
	}

	return btcutil.Amount(result.Value)
}

// SetExternalKeyIDs stores the given list of external key IDs
func (k Keeper) SetExternalKeyIDs(ctx sdk.Context, keyIDs []tss.KeyID) {
	values := make([]*gogoprototypes.Value, len(keyIDs))
	for i, keyID := range keyIDs {
		values[i] = &gogoprototypes.Value{
			Kind: &gogoprototypes.Value_StringValue{StringValue: string(keyID)},
		}
	}

	k.getStore(ctx).Set(externalKeyIDsKey, &gogoprototypes.ListValue{Values: values})
}

// GetExternalKeyIDs retrieves the current list of external key IDs
func (k Keeper) GetExternalKeyIDs(ctx sdk.Context) ([]tss.KeyID, bool) {
	var listValue gogoprototypes.ListValue
	if !k.getStore(ctx).Get(externalKeyIDsKey, &listValue) {
		return nil, false
	}

	keyIDs := make([]tss.KeyID, len(listValue.Values))
	for i, value := range listValue.Values {
		keyIDs[i] = tss.KeyID(value.GetStringValue())
	}

	return keyIDs, true
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
