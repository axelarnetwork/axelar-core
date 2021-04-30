package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const (
	pendingOutpointPrefix   = "pend_"
	confirmedOutPointPrefix = "conf_"
	spentOutPointPrefix     = "spent_"
	addrPrefix              = "addr_"
	dustAmtPrefix           = "dust_"

	unsignedTxKey          = "unsignedTx"
	signedTxKey            = "signedTx"
	masterKeyUtxoExistsKey = "master_key_utxo_exists"
)

var _ types.BTCKeeper = Keeper{}

// Keeper provides access to all state changes regarding the Bitcoin module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.LegacyAmino
	params   params.Subspace
}

// NewKeeper returns a new keeper object
func NewKeeper(cdc *codec.LegacyAmino, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
	return Keeper{cdc: cdc, storeKey: storeKey, params: paramSpace.WithKeyTable(types.KeyTable())}
}

// SetParams sets the bitcoin module's parameters
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) {
	k.params.SetParamSet(ctx, &p)
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

// Codec returns the codec used by the keeper to marshal and unmarshal data
func (k Keeper) Codec() *codec.LegacyAmino {
	return k.cdc
}

// SetAddress stores the given address information
func (k Keeper) SetAddress(ctx sdk.Context, address types.AddressInfo) {
	// btcutil.Address (and it's implementations) can't be serialized with amino,
	// so we use a helper struct to get around that problem
	a := struct {
		Addr   string
		Script types.RedeemScript
		Key    tss.Key
	}{
		Addr:   address.EncodeAddress(),
		Script: address.RedeemScript,
		Key:    address.Key,
	}
	ctx.KVStore(k.storeKey).Set([]byte(addrPrefix+address.EncodeAddress()), k.Codec().MustMarshalBinaryLengthPrefixed(a))
}

// GetAddress returns the address information for the given encoded address
func (k Keeper) GetAddress(ctx sdk.Context, encodedAddress string) (types.AddressInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(addrPrefix + encodedAddress))
	if bz == nil {
		return types.AddressInfo{}, false
	}

	// btcutil.Address (and it's implementations) can't be serialized with amino,
	// so we use a helper struct to get around that problem
	var a struct {
		Addr   string
		Script types.RedeemScript
		Key    tss.Key
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &a)
	addr, _ := btcutil.DecodeAddress(a.Addr, k.GetNetwork(ctx).Params())
	return types.AddressInfo{
		Address:      addr,
		RedeemScript: a.Script,
		Key:          a.Key,
	}, true
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

	prefixes := map[types.OutPointState]string{
		types.CONFIRMED: confirmedOutPointPrefix,
		types.SPENT:     spentOutPointPrefix,
	}

	// the order of this iteration is non-deterministic,
	// the only reason this is correct is because exactly one of the cases is true
	for state, prefix := range prefixes {
		bz := ctx.KVStore(k.storeKey).Get([]byte(prefix + outPoint.String()))
		if bz != nil {
			k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
			return info, state, true
		}

	}

	return types.OutPointInfo{}, 0, false
}

// SetPendingOutpointInfo stores an unconfirmed outpoint.
// Since the information is not yet confirmed the outpoint info is not necessarily unique.
// Therefore we need to store by the poll that confirms/rejects it
func (k Keeper) SetPendingOutpointInfo(ctx sdk.Context, poll exported.PollMeta, info types.OutPointInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(pendingOutpointPrefix+poll.String()), bz)
}

// DeletePendingOutPointInfo deletes the outpoint information associated with the given poll
func (k Keeper) DeletePendingOutPointInfo(ctx sdk.Context, poll exported.PollMeta) {
	ctx.KVStore(k.storeKey).Delete([]byte(pendingOutpointPrefix + poll.String()))
}

// SetOutpointInfo stores confirmed or spent outpoints
func (k Keeper) SetOutpointInfo(ctx sdk.Context, info types.OutPointInfo, state types.OutPointState) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
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
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(unsignedTxKey), bz)
}

// GetUnsignedTx returns the raw unsigned transaction for outpoint consolidation
func (k Keeper) GetUnsignedTx(ctx sdk.Context) (*wire.MsgTx, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(unsignedTxKey))
	if bz == nil {
		return nil, false
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx, true
}

// DeleteUnsignedTx deletes the raw unsigned transaction for outpoint consolidation
func (k Keeper) DeleteUnsignedTx(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete([]byte(unsignedTxKey))
}

// SetSignedTx stores the signed transaction for outpoint consolidation
func (k Keeper) SetSignedTx(ctx sdk.Context, tx *wire.MsgTx) {
	ctx.KVStore(k.storeKey).Set([]byte(signedTxKey), k.cdc.MustMarshalBinaryLengthPrefixed(tx))
}

// GetSignedTx returns the signed transaction for outpoint consolidation
func (k Keeper) GetSignedTx(ctx sdk.Context) (*wire.MsgTx, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(signedTxKey))
	if bz == nil {
		return nil, false
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	return tx, true
}

// DeleteSignedTx deletes the signed transaction for outpoint consolidation
func (k Keeper) DeleteSignedTx(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete([]byte(signedTxKey))
}

// SetDustAmount stores the dust amount for a destination bitcoin address
func (k Keeper) SetDustAmount(ctx sdk.Context, encodedAddress string, amount btcutil.Amount) {
	ctx.KVStore(k.storeKey).Set([]byte(dustAmtPrefix+encodedAddress), k.cdc.MustMarshalBinaryLengthPrefixed(amount))
}

// GetDustAmount returns the dust amount for a destination bitcoin address
func (k Keeper) GetDustAmount(ctx sdk.Context, encodedAddress string) btcutil.Amount {

	bz := ctx.KVStore(k.storeKey).Get([]byte(dustAmtPrefix + encodedAddress))
	if bz == nil {
		return 0
	}
	var amt btcutil.Amount
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &amt)
	return amt
}

// DeleteDustAmount deletes the dust amount for a destination bitcoin address
func (k Keeper) DeleteDustAmount(ctx sdk.Context, encodedAddress string) {
	ctx.KVStore(k.storeKey).Delete([]byte(dustAmtPrefix + encodedAddress))
}

// SetMasterKeyOutpointExists sets existence for UTXO controlled by the master key
func (k Keeper) SetMasterKeyOutpointExists(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Set([]byte(masterKeyUtxoExistsKey), []byte{})
}

// DoesMasterKeyOutpointExist returns true if there is any UTXO controlled by the master key; otherwise, false
func (k Keeper) DoesMasterKeyOutpointExist(ctx sdk.Context) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(masterKeyUtxoExistsKey))
}
