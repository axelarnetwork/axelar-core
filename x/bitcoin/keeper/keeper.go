package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	rawKey = "rawTx"

	unverifiedOutpointPrefix = "unver_"
	verifiedOutPointPrefix   = "ver_"
	spentOutPointPrefix      = "spent_"
	scriptPrefix             = "script_"
	keyIDByAddrPrefix        = "addrID_"
	keyIDByOutPointPrefix    = "outID_"
	signedTxKey              = "signedTx_"
)

// Keeper provides access to all state changes regarding the Bitcoin module
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	params   params.Subspace
}

// NewKeeper returns a new keeper object
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
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
// before axelar will accept it for verification.
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

// Codec returns the codec used by the keeper to marshal and unmarshal data
func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

// SetKeyIDByAddress stores the ID of the key that controls the given address
func (k Keeper) SetKeyIDByAddress(ctx sdk.Context, address btcutil.Address, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keyIDByAddrPrefix+address.String()), []byte(keyID))
}

// GetKeyIDByAddress returns the ID of the key that was used to create the given address
func (k Keeper) GetKeyIDByAddress(ctx sdk.Context, address string) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keyIDByAddrPrefix + address))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// SetKeyIDByOutpoint stores the ID of the key that controls the address corresponding to the given outpoint
func (k Keeper) SetKeyIDByOutpoint(ctx sdk.Context, outpoint *wire.OutPoint, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keyIDByOutPointPrefix+outpoint.String()), []byte(keyID))
}

// GetKeyIDByOutpoint returns the ID of the key that controls the address corresponding to the given outpoint
func (k Keeper) GetKeyIDByOutpoint(ctx sdk.Context, outpoint *wire.OutPoint) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keyIDByOutPointPrefix + outpoint.String()))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// GetVerifiedOutPointInfo returns additional information for the given outpoint, if it was verified
func (k Keeper) GetVerifiedOutPointInfo(ctx sdk.Context, outPoint *wire.OutPoint) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(verifiedOutPointPrefix + outPoint.String()))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var out types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &out)

	return out, true
}

// SetUnverifiedOutpointInfo stores the outpoint information of an unverified Bitcoin transaction
func (k Keeper) SetUnverifiedOutpointInfo(ctx sdk.Context, info types.OutPointInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(unverifiedOutpointPrefix+info.OutPoint.String()), bz)
}

// GetUnverifiedOutPointInfo returns additional information for the given unverified outpoint
func (k Keeper) GetUnverifiedOutPointInfo(ctx sdk.Context, outpoint *wire.OutPoint) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(unverifiedOutpointPrefix + outpoint.String()))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var info types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)

	return info, true
}

// ProcessVerificationResult stores the info related to the specified outpoint (format txID:voutIdx) permanently if confirmed or discards the data otherwise.
// Does nothing if the outPoint is unknown
func (k Keeper) ProcessVerificationResult(ctx sdk.Context, outPoint *wire.OutPoint, verified bool) {
	outStr := outPoint.String()
	bz := ctx.KVStore(k.storeKey).Get([]byte(unverifiedOutpointPrefix + outStr))
	if bz == nil {
		k.Logger(ctx).Debug(fmt.Sprintf("outpoint %s is not known", outPoint))
		return
	}
	ctx.KVStore(k.storeKey).Delete([]byte(unverifiedOutpointPrefix + outStr))
	if verified {
		ctx.KVStore(k.storeKey).Set([]byte(verifiedOutPointPrefix+outStr), bz)
	}
}

// SetRedeemScriptByAddress stores the full redeem script corresponding to the given address (the hash of the script was used to generate the address)
func (k Keeper) SetRedeemScriptByAddress(ctx sdk.Context, address btcutil.Address, script []byte) {
	ctx.KVStore(k.storeKey).Set([]byte(scriptPrefix+address.String()), script)
}

// GetRedeemScript returns the full redeem script corresponding to the given address (the hash of the script was used to generate the address)
func (k Keeper) GetRedeemScript(ctx sdk.Context, address btcutil.Address) ([]byte, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(scriptPrefix + address.String()))
	return bz, bz != nil
}

// GetHashesToSign returns the hash that needs to be signed to create a valid signature for the given unsigned Bitcoin transaction
func (k Keeper) GetHashesToSign(ctx sdk.Context, rawTx *wire.MsgTx) ([][]byte, error) {
	var hashes [][]byte
	for i, in := range rawTx.TxIn {
		prevOutInfo, ok := k.GetSpentOutPointInfo(ctx, &in.PreviousOutPoint)
		if !ok {
			return nil, fmt.Errorf("transaction ID is not known")
		}

		addr, err := btcutil.DecodeAddress(prevOutInfo.Address, k.GetNetwork(ctx).Params)
		if err != nil {
			return nil, err
		}

		script, ok := k.GetRedeemScript(ctx, addr)
		if !ok {
			return nil, fmt.Errorf("could not find a redeem script for outpoint %s", in.PreviousOutPoint.String())
		}
		hash, err := txscript.CalcWitnessSigHash(script, txscript.NewTxSigHashes(rawTx), txscript.SigHashAll, rawTx, i, int64(prevOutInfo.Amount))
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

// AssembleBtcTx assembles the unsigned transaction and given signature.
// Returns a an error the resulting signed Bitcoin transaction is invalid.
func (k Keeper) AssembleBtcTx(ctx sdk.Context, rawTx *wire.MsgTx, sigs []btcec.Signature) (*wire.MsgTx, error) {
	for i, sig := range sigs {
		prevOutInfo, ok := k.GetSpentOutPointInfo(ctx, &rawTx.TxIn[i].PreviousOutPoint)
		if !ok {
			return nil, fmt.Errorf("transaction ID is not known")
		}

		addr, err := btcutil.DecodeAddress(prevOutInfo.Address, k.GetNetwork(ctx).Params)
		if err != nil {
			return nil, err
		}

		redeemScript, ok := k.GetRedeemScript(ctx, addr)
		if !ok {
			return nil, fmt.Errorf("redeem script for address %s not found", addr.EncodeAddress())
		}

		rawTx.TxIn[i].Witness = types.CreateTxWitness(sig, redeemScript)

		payScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}

		if err := types.ValidateTxScript(rawTx, i, prevOutInfo, payScript); err != nil {
			return nil, err
		}
	}

	return rawTx, nil
}

// GetNetwork returns the connected Bitcoin network (main, test, regtest)
func (k Keeper) GetNetwork(ctx sdk.Context) types.Network {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return network
}

// GetVerifiedOutPointInfos returns information about all unspent verified outpoints controlled by Axelar-Core
func (k Keeper) GetVerifiedOutPointInfos(ctx sdk.Context) []types.OutPointInfo {
	var outs []types.OutPointInfo
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(verifiedOutPointPrefix))
	for ; iter.Valid(); iter.Next() {
		var info types.OutPointInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &info)
		outs = append(outs, info)
	}
	return outs
}

// SpendVerifiedOutPoint marks the given outpoint as spent
func (k Keeper) SpendVerifiedOutPoint(ctx sdk.Context, outPoint string) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(verifiedOutPointPrefix + outPoint))
	if bz == nil {
		k.Logger(ctx).Debug(fmt.Sprintf("outpoint %s is either unknown, unverified or already spent", outPoint))
		return
	}
	ctx.KVStore(k.storeKey).Delete([]byte(verifiedOutPointPrefix + outPoint))
	ctx.KVStore(k.storeKey).Set([]byte(spentOutPointPrefix+outPoint), bz)
}

// GetSpentOutPointInfo returns additional information for the given outpoint, if it was verified and used as a transaction input
func (k Keeper) GetSpentOutPointInfo(ctx sdk.Context, outPoint *wire.OutPoint) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(spentOutPointPrefix + outPoint.String()))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var out types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &out)

	return out, true
}

// SetRawTx stores a raw transaction for outpoint consolidation
func (k Keeper) SetRawTx(ctx sdk.Context, tx *wire.MsgTx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawKey), bz)
}

// GetRawTx returns the raw unsigned transaction for outpoint consolidation
func (k Keeper) GetRawTx(ctx sdk.Context) *wire.MsgTx {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawKey))
	if bz == nil {
		return nil
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx
}

// DeleteRawTx deletes the raw unsigned transaction for outpoint consolidation
func (k Keeper) DeleteRawTx(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete([]byte(rawKey))
}

// SetSignedTx stores the signed transaction for outpoint consolidation
func (k Keeper) SetSignedTx(ctx sdk.Context, tx *wire.MsgTx) {
	ctx.KVStore(k.storeKey).Set([]byte(signedTxKey), k.cdc.MustMarshalBinaryLengthPrefixed(tx))
}

// GetSignedTx returns the signed transaction for outpoint consolidation
func (k Keeper) GetSignedTx(ctx sdk.Context) *wire.MsgTx {
	bz := ctx.KVStore(k.storeKey).Get([]byte(signedTxKey))
	if bz == nil {
		return nil
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)
	return tx
}

// DeleteSignedTx deletes the signed transaction for outpoint consolidation
func (k Keeper) DeleteSignedTx(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete([]byte(signedTxKey))
}
