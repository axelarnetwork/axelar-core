package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const (
	pendingOutpointPrefix   = "pend_"
	confirmedOutPointPrefix = "conf_"
	spentOutPointPrefix     = "spent_"
	addrPrefix              = "addr_"
	dustAmtPrefix           = "dust_"

	anyoneCanSpendAddressKey = "anyone_can_spend_address"
	unsignedTxKey            = "unsignedTx"
	signedTxKey              = "signedTx"
	masterKeyVoutKey         = "master_key_vout"

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
	k.getStore(ctx).Set(utils.LowerCaseKey(anyoneCanSpendAddressKey), &anyoneCanSpendAddress)
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
	ok := k.getStore(ctx).Get(utils.LowerCaseKey(anyoneCanSpendAddressKey), &address)
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

// GetMinimumWithdrawalAmount returns the minimum withdrawal threshold
func (k Keeper) GetMinimumWithdrawalAmount(ctx sdk.Context) btcutil.Amount {
	var result btcutil.Amount
	k.params.Get(ctx, types.KeyMinimumWithdrawalAmount, &result)

	return result
}

// GetMaxInputCount returns the max input count
func (k Keeper) GetMaxInputCount(ctx sdk.Context) int64 {
	var result int64
	k.params.Get(ctx, types.KeyMaxInputCount, &result)

	return result
}

// SetAddress stores the given address information
func (k Keeper) SetAddress(ctx sdk.Context, address types.AddressInfo) {
	k.getStore(ctx).Set(utils.LowerCaseKey(address.Address).WithPrefix(addrPrefix), &address)
}

// GetAddress returns the address information for the given encoded address
func (k Keeper) GetAddress(ctx sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
	var address types.AddressInfo
	ok := k.getStore(ctx).Get(utils.LowerCaseKey(encodedAddress).WithPrefix(addrPrefix), &address)
	return address, ok
}

// DeleteOutpointInfo deletes a the given outpoint if known
func (k Keeper) DeleteOutpointInfo(ctx sdk.Context, outPoint wire.OutPoint) {
	// delete is a noop if key does not exist
	key := utils.ToLowerCaseKey(outPoint)
	k.getStore(ctx).Delete(key.WithPrefix(confirmedOutPointPrefix))
	k.getStore(ctx).Delete(key.WithPrefix(spentOutPointPrefix))
}

// GetPendingOutPointInfo returns outpoint information associated with the given poll
func (k Keeper) GetPendingOutPointInfo(ctx sdk.Context, poll exported.PollMeta) (types.OutPointInfo, bool) {
	var info types.OutPointInfo
	ok := k.getStore(ctx).Get(utils.ToLowerCaseKey(poll).WithPrefix(pendingOutpointPrefix), &info)
	return info, ok
}

// GetOutPointInfo returns additional information for the given outpoint
func (k Keeper) GetOutPointInfo(ctx sdk.Context, outPoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
	var info types.OutPointInfo

	key := utils.ToLowerCaseKey(outPoint)
	ok := k.getStore(ctx).Get(key.WithPrefix(confirmedOutPointPrefix), &info)
	if ok {
		return info, types.CONFIRMED, true
	}

	ok = k.getStore(ctx).Get(key.WithPrefix(spentOutPointPrefix), &info)
	if ok {
		return info, types.SPENT, true
	}

	return types.OutPointInfo{}, 0, false
}

// SetPendingOutpointInfo stores an unconfirmed outpoint.
// Since the information is not yet confirmed the outpoint info is not necessarily unique.
// Therefore we need to store by the poll that confirms/rejects it
func (k Keeper) SetPendingOutpointInfo(ctx sdk.Context, poll exported.PollMeta, info types.OutPointInfo) {
	k.getStore(ctx).Set(utils.ToLowerCaseKey(poll).WithPrefix(pendingOutpointPrefix), &info)
}

// DeletePendingOutPointInfo deletes the outpoint information associated with the given poll
func (k Keeper) DeletePendingOutPointInfo(ctx sdk.Context, poll exported.PollMeta) {
	k.getStore(ctx).Delete(utils.ToLowerCaseKey(poll).WithPrefix(pendingOutpointPrefix))
}

// SetOutpointInfo stores confirmed or spent outpoints
func (k Keeper) SetOutpointInfo(ctx sdk.Context, info types.OutPointInfo, state types.OutPointState) {
	key := utils.LowerCaseKey(info.OutPoint)
	switch state {
	case types.CONFIRMED:
		k.GetConfirmedOutpointInfoQueue(ctx).Enqueue(key.WithPrefix(confirmedOutPointPrefix), &info)
	case types.SPENT:
		k.getStore(ctx).Set(key.WithPrefix(spentOutPointPrefix), &info)
	default:
		panic("invalid outpoint state")
	}
}

// GetConfirmedOutpointInfoQueue returns the queue for confirmed outpoint infos
func (k Keeper) GetConfirmedOutpointInfoQueue(ctx sdk.Context) utils.KVQueue {
	return utils.NewBlockHeightKVQueue(k.getStore(ctx), ctx, confirmedOutpointQueueName)
}

// SetUnsignedTx stores a raw transaction for outpoint consolidation
func (k Keeper) SetUnsignedTx(ctx sdk.Context, tx *wire.MsgTx) {
	k.getStore(ctx).SetRaw(utils.LowerCaseKey(unsignedTxKey), types.MustEncodeTx(tx))
}

// GetUnsignedTx returns the raw unsigned transaction for outpoint consolidation
func (k Keeper) GetUnsignedTx(ctx sdk.Context) (*wire.MsgTx, bool) {
	bz := k.getStore(ctx).GetRaw(utils.LowerCaseKey(unsignedTxKey))
	if bz == nil {
		return nil, false
	}

	tx := types.MustDecodeTx(bz)

	return &tx, true
}

// DeleteUnsignedTx deletes the raw unsigned transaction for outpoint consolidation
func (k Keeper) DeleteUnsignedTx(ctx sdk.Context) {
	k.getStore(ctx).Delete(utils.LowerCaseKey(unsignedTxKey))
}

// SetSignedTx stores the signed transaction for outpoint consolidation
func (k Keeper) SetSignedTx(ctx sdk.Context, tx *wire.MsgTx) {
	k.getStore(ctx).SetRaw(utils.LowerCaseKey(signedTxKey), types.MustEncodeTx(tx))
}

// GetSignedTx returns the signed transaction for outpoint consolidation
func (k Keeper) GetSignedTx(ctx sdk.Context) (*wire.MsgTx, bool) {
	bz := k.getStore(ctx).GetRaw(utils.LowerCaseKey(signedTxKey))
	if bz == nil {
		return nil, false
	}

	tx := types.MustDecodeTx(bz)

	return &tx, true
}

// DeleteSignedTx deletes the signed transaction for outpoint consolidation
func (k Keeper) DeleteSignedTx(ctx sdk.Context) {
	k.getStore(ctx).Delete(utils.LowerCaseKey(signedTxKey))
}

// SetDustAmount stores the dust amount for a destination bitcoin address
func (k Keeper) SetDustAmount(ctx sdk.Context, encodedAddress string, amount btcutil.Amount) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(amount))

	k.getStore(ctx).SetRaw(utils.LowerCaseKey(encodedAddress).WithPrefix(dustAmtPrefix), bz)
}

// GetDustAmount returns the dust amount for a destination bitcoin address
func (k Keeper) GetDustAmount(ctx sdk.Context, encodedAddress string) btcutil.Amount {
	bz := k.getStore(ctx).GetRaw(utils.LowerCaseKey(encodedAddress).WithPrefix(dustAmtPrefix))
	if bz == nil {
		return 0
	}

	return btcutil.Amount(int64(binary.LittleEndian.Uint64(bz)))
}

// DeleteDustAmount deletes the dust amount for a destination bitcoin address
func (k Keeper) DeleteDustAmount(ctx sdk.Context, encodedAddress string) {
	k.getStore(ctx).Delete(utils.LowerCaseKey(encodedAddress).WithPrefix(dustAmtPrefix))
}

// SetMasterKeyVout sets the index of the consolidation outpoint
func (k Keeper) SetMasterKeyVout(ctx sdk.Context, vout uint32) {
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, vout)
	k.getStore(ctx).SetRaw(utils.LowerCaseKey(masterKeyVoutKey), bz)
}

// GetMasterKeyVout returns the index of the consolidation outpoint if there is any UTXO controlled by the master key; otherwise, false
func (k Keeper) GetMasterKeyVout(ctx sdk.Context) (uint32, bool) {
	bz := k.getStore(ctx).GetRaw(utils.LowerCaseKey(masterKeyVoutKey))
	if bz == nil {
		return 0, false
	}
	return binary.LittleEndian.Uint32(bz), true
}

func (k Keeper) getStore(ctx sdk.Context) utils.NormalizedKVStore {
	return utils.NewNormalizedStore(ctx.KVStore(k.storeKey), k.cdc)
}
