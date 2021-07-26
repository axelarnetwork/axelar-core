package keeper

import (
	"encoding/binary"
	"fmt"

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
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	pendingOutpointPrefix    = utils.KeyFromStr("pend_")
	confirmedOutPointPrefix  = utils.KeyFromStr("conf_")
	spentOutPointPrefix      = utils.KeyFromStr("spent_")
	addrPrefix               = utils.KeyFromStr("addr_")
	dustAmtPrefix            = utils.KeyFromStr("dust_")
	signedTxPrefix           = utils.KeyFromStr("signed_tx_")
	anyoneCanSpendVoutPrefix = utils.KeyFromStr("anyone_can_spend_vout_")

	anyoneCanSpendAddressKey = utils.KeyFromStr("anyone_can_spend_address")
	unsignedTxKey            = utils.KeyFromStr("unsigned_tx")
	latestSignedTxHashKey    = utils.KeyFromStr("latest_signed_tx_hash")

	confirmedOutpointQueueName = "confirmed_outpoint"
)

var _ types.BTCKeeper = Keeper{}

// Keeper provides access to all state changes regarding the Bitcoin module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler
	params   params.Subspace
}

// NewKeeper returns a new keeper object
func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
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
func (k Keeper) GetPendingOutPointInfo(ctx sdk.Context, key exported.PollKey) (types.OutPointInfo, bool) {
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
		return info, types.CONFIRMED, true
	}

	ok = k.getStore(ctx).Get(spentOutPointPrefix.Append(key), &info)
	if ok {
		return info, types.SPENT, true
	}

	return types.OutPointInfo{}, 0, false
}

// SetPendingOutpointInfo stores an unconfirmed outpoint.
// Since the information is not yet confirmed the outpoint info is not necessarily unique.
// Therefore we need to store by the poll that confirms/rejects it
func (k Keeper) SetPendingOutpointInfo(ctx sdk.Context, key exported.PollKey, info types.OutPointInfo) {
	k.getStore(ctx).Set(pendingOutpointPrefix.Append(utils.LowerCaseKey(key.String())), &info)
}

// DeletePendingOutPointInfo deletes the outpoint information associated with the given poll
func (k Keeper) DeletePendingOutPointInfo(ctx sdk.Context, key exported.PollKey) {
	k.getStore(ctx).Delete(pendingOutpointPrefix.Append(utils.LowerCaseKey(key.String())))
}

// SetSpentOutpointInfo stores the given outpoint info as spent
func (k Keeper) SetSpentOutpointInfo(ctx sdk.Context, info types.OutPointInfo) {
	key := utils.LowerCaseKey(info.OutPoint)

	k.getStore(ctx).Set(spentOutPointPrefix.Append(key), &info)
}

// SetConfirmedOutpointInfo stores the given outpoint info as confirmed and push it into the queue of given keyID
func (k Keeper) SetConfirmedOutpointInfo(ctx sdk.Context, keyID string, info types.OutPointInfo) {
	key := utils.LowerCaseKey(info.OutPoint)

	k.GetConfirmedOutpointInfoQueueForKey(ctx, keyID).Enqueue(confirmedOutPointPrefix.Append(key), &info)
}

// GetConfirmedOutpointInfoQueueForKey retrieves the outpoint info queue for the given keyID
func (k Keeper) GetConfirmedOutpointInfoQueueForKey(ctx sdk.Context, keyID string) utils.KVQueue {
	queueName := fmt.Sprintf("%s_%s", confirmedOutpointQueueName, keyID)

	return utils.NewBlockHeightKVQueue(queueName, k.getStore(ctx), ctx.BlockHeight(), k.Logger(ctx))
}

// SetUnsignedTx stores a raw transaction for outpoint consolidation
func (k Keeper) SetUnsignedTx(ctx sdk.Context, tx *types.Transaction) {
	k.getStore(ctx).Set(unsignedTxKey, tx)
}

// GetUnsignedTx returns the raw unsigned transaction for outpoint consolidation
func (k Keeper) GetUnsignedTx(ctx sdk.Context) (*types.Transaction, bool) {
	var result types.Transaction
	if ok := k.getStore(ctx).Get(unsignedTxKey, &result); !ok {
		return nil, false
	}

	return &result, true
}

// DeleteUnsignedTx deletes the raw unsigned transaction for outpoint consolidation
func (k Keeper) DeleteUnsignedTx(ctx sdk.Context) {
	k.getStore(ctx).Delete(unsignedTxKey)
}

// SetSignedTx stores the signed transaction for outpoint consolidation
func (k Keeper) SetSignedTx(ctx sdk.Context, tx *wire.MsgTx) {
	txHash := tx.TxHash()

	k.getStore(ctx).SetRaw(latestSignedTxHashKey, txHash.CloneBytes())
	k.getStore(ctx).SetRaw(signedTxPrefix.Append(utils.LowerCaseKey(txHash.String())), types.MustEncodeTx(tx))
}

// GetSignedTx returns the signed transaction for outpoint consolidation
// TODO: think about how to get all signed txs in the correct order
func (k Keeper) GetSignedTx(ctx sdk.Context, txHash chainhash.Hash) (*wire.MsgTx, bool) {
	bz := k.getStore(ctx).GetRaw(signedTxPrefix.Append(utils.LowerCaseKey(txHash.String())))
	if bz == nil {
		return nil, false
	}

	tx := types.MustDecodeTx(bz)

	return &tx, true
}

// GetLatestSignedTxHash retrieves the tx hash of the most recent signed transaction
func (k Keeper) GetLatestSignedTxHash(ctx sdk.Context) (*chainhash.Hash, bool) {
	bz := k.getStore(ctx).GetRaw(latestSignedTxHashKey)
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

// GetAnyoneCanSpendVout retrieves the vout of anyone-can-spend output of given transaction hash
func (k Keeper) GetAnyoneCanSpendVout(ctx sdk.Context, txHash chainhash.Hash) (int64, bool) {
	var result gogoprototypes.Int64Value
	if ok := k.getStore(ctx).Get(anyoneCanSpendVoutPrefix.Append(utils.LowerCaseKey(txHash.String())), &result); !ok {
		return 0, false
	}

	return result.Value, true
}

// SetAnyoneCanSpendVout sets the vout of anyone-can-spend output for the given transaction hash
func (k Keeper) SetAnyoneCanSpendVout(ctx sdk.Context, txHash chainhash.Hash, vout int64) {
	k.getStore(ctx).Set(anyoneCanSpendVoutPrefix.Append(utils.LowerCaseKey(txHash.String())), &gogoprototypes.Int64Value{Value: int64(vout)})
}

func (k Keeper) getStore(ctx sdk.Context) utils.KVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
