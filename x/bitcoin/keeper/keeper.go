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
	k.setAddress(ctx, anyoneCanSpendAddressKey, types.NewAnyoneCanSpendAddress(p.Network))
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
	address, found := k.getAddress(ctx, anyoneCanSpendAddressKey)
	if !found {
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

// SetAddress stores the given address information
func (k Keeper) SetAddress(ctx sdk.Context, address types.AddressInfo) {
	k.setAddress(ctx, addrPrefix+address.Address, address)
}

func (k Keeper) setAddress(ctx sdk.Context, key string, address types.AddressInfo) {
	ctx.KVStore(k.storeKey).Set([]byte(key), k.cdc.MustMarshalBinaryLengthPrefixed(&address))
}

// GetAddress returns the address information for the given encoded address
func (k Keeper) GetAddress(ctx sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
	return k.getAddress(ctx, addrPrefix+encodedAddress)
}

func (k Keeper) getAddress(ctx sdk.Context, key string) (types.AddressInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(key))
	if bz == nil {
		return types.AddressInfo{}, false
	}

	var address types.AddressInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &address)

	return address, true
}

// DeleteOutpointInfo deletes a the given outpoint if known
func (k Keeper) DeleteOutpointInfo(ctx sdk.Context, outPoint wire.OutPoint) {
	// delete is a noop if key does not exist
	ctx.KVStore(k.storeKey).Delete([]byte(confirmedOutPointPrefix + outPoint.String()))
	ctx.KVStore(k.storeKey).Delete([]byte(spentOutPointPrefix + outPoint.String()))
}

// GetPendingOutPointInfo returns outpoint information associated with the given poll
func (k Keeper) GetPendingOutPointInfo(ctx sdk.Context, poll exported.PollMeta) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingOutpointPrefix + poll.String()))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var info types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
	return info, true
}

// GetOutPointInfo returns additional information for the given outpoint
func (k Keeper) GetOutPointInfo(ctx sdk.Context, outPoint wire.OutPoint) (types.OutPointInfo, types.OutPointState, bool) {
	var info types.OutPointInfo

	bz := ctx.KVStore(k.storeKey).Get([]byte(confirmedOutPointPrefix + outPoint.String()))
	if bz != nil {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
		return info, types.CONFIRMED, true
	}

	bz = ctx.KVStore(k.storeKey).Get([]byte(spentOutPointPrefix + outPoint.String()))
	if bz != nil {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
		return info, types.SPENT, true
	}

	return types.OutPointInfo{}, 0, false
}

// SetPendingOutpointInfo stores an unconfirmed outpoint.
// Since the information is not yet confirmed the outpoint info is not necessarily unique.
// Therefore we need to store by the poll that confirms/rejects it
func (k Keeper) SetPendingOutpointInfo(ctx sdk.Context, poll exported.PollMeta, info types.OutPointInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&info)
	ctx.KVStore(k.storeKey).Set([]byte(pendingOutpointPrefix+poll.String()), bz)
}

// DeletePendingOutPointInfo deletes the outpoint information associated with the given poll
func (k Keeper) DeletePendingOutPointInfo(ctx sdk.Context, poll exported.PollMeta) {
	ctx.KVStore(k.storeKey).Delete([]byte(pendingOutpointPrefix + poll.String()))
}

// SetOutpointInfo stores confirmed or spent outpoints
func (k Keeper) SetOutpointInfo(ctx sdk.Context, info types.OutPointInfo, state types.OutPointState) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&info)

	switch state {
	case types.CONFIRMED:
		ctx.KVStore(k.storeKey).Set([]byte(confirmedOutPointPrefix+info.OutPoint), bz)
	case types.SPENT:
		ctx.KVStore(k.storeKey).Set([]byte(spentOutPointPrefix+info.OutPoint), bz)
	default:
		panic("invalid outpoint state")
	}
}

// GetConfirmedOutPointInfos returns information about all confirmed outpoints
func (k Keeper) GetConfirmedOutPointInfos(ctx sdk.Context) []types.OutPointInfo {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(confirmedOutPointPrefix))

	var outs []types.OutPointInfo
	for ; iter.Valid(); iter.Next() {
		var info types.OutPointInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &info)
		outs = append(outs, info)
	}
	return outs
}

// SetUnsignedTx stores a raw transaction for outpoint consolidation
func (k Keeper) SetUnsignedTx(ctx sdk.Context, tx *wire.MsgTx) {
	ctx.KVStore(k.storeKey).Set([]byte(unsignedTxKey), types.MustEncodeTx(tx))
}

// GetUnsignedTx returns the raw unsigned transaction for outpoint consolidation
func (k Keeper) GetUnsignedTx(ctx sdk.Context) (*wire.MsgTx, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(unsignedTxKey))
	if bz == nil {
		return nil, false
	}

	tx := types.MustDecodeTx(bz)

	return &tx, true
}

// DeleteUnsignedTx deletes the raw unsigned transaction for outpoint consolidation
func (k Keeper) DeleteUnsignedTx(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete([]byte(unsignedTxKey))
}

// SetSignedTx stores the signed transaction for outpoint consolidation
func (k Keeper) SetSignedTx(ctx sdk.Context, tx *wire.MsgTx) {
	ctx.KVStore(k.storeKey).Set([]byte(signedTxKey), types.MustEncodeTx(tx))
}

// GetSignedTx returns the signed transaction for outpoint consolidation
func (k Keeper) GetSignedTx(ctx sdk.Context) (*wire.MsgTx, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(signedTxKey))
	if bz == nil {
		return nil, false
	}

	tx := types.MustDecodeTx(bz)

	return &tx, true
}

// DeleteSignedTx deletes the signed transaction for outpoint consolidation
func (k Keeper) DeleteSignedTx(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete([]byte(signedTxKey))
}

// SetDustAmount stores the dust amount for a destination bitcoin address
func (k Keeper) SetDustAmount(ctx sdk.Context, encodedAddress string, amount btcutil.Amount) {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(amount))

	ctx.KVStore(k.storeKey).Set([]byte(dustAmtPrefix+encodedAddress), bz)
}

// GetDustAmount returns the dust amount for a destination bitcoin address
func (k Keeper) GetDustAmount(ctx sdk.Context, encodedAddress string) btcutil.Amount {
	bz := ctx.KVStore(k.storeKey).Get([]byte(dustAmtPrefix + encodedAddress))
	if bz == nil {
		return 0
	}

	return btcutil.Amount(int64(binary.LittleEndian.Uint64(bz)))
}

// DeleteDustAmount deletes the dust amount for a destination bitcoin address
func (k Keeper) DeleteDustAmount(ctx sdk.Context, encodedAddress string) {
	ctx.KVStore(k.storeKey).Delete([]byte(dustAmtPrefix + encodedAddress))
}

// SetMasterKeyVout sets the index of the consolidation outpoint
func (k Keeper) SetMasterKeyVout(ctx sdk.Context, vout uint32) {
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, vout)
	ctx.KVStore(k.storeKey).Set([]byte(masterKeyVoutKey), bz)
}

// GetMasterKeyVout returns the index of the consolidation outpoint if there is any UTXO controlled by the master key; otherwise, false
func (k Keeper) GetMasterKeyVout(ctx sdk.Context) (uint32, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(masterKeyVoutKey))
	if bz == nil {
		return 0, false
	}
	return binary.LittleEndian.Uint32(bz), true
}
