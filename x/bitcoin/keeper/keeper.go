package keeper

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"
)

const (
	rawPrefix             = "raw_"
	outPointPrefix        = "out_"
	pendingPrefix         = "pend_"
	addrPrefix            = "addr_"
	scriptPrefix          = "script_"
	keyIDbyAddrPrefix     = "addrID_"
	keyIDbyOutPointPrefix = "outID_"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	params   params.Subspace
}

func NewBtcKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace) Keeper {
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

func (k Keeper) SetTrackedAddress(ctx sdk.Context, address string) {
	ctx.KVStore(k.storeKey).Set([]byte(addrPrefix+address), []byte{})
}

func (k Keeper) GetTrackedAddress(ctx sdk.Context, address string) string {
	val := ctx.KVStore(k.storeKey).Get([]byte(addrPrefix + address))
	if val == nil {
		return ""
	}
	return address
}

// GetRequiredConfirmationHeight returns the minimum number of confirmations a transaction must have on Bitcoin
// before axelar will accept it for verification.
func (k Keeper) GetRequiredConfirmationHeight(ctx sdk.Context) uint64 {
	var h uint64
	k.params.Get(ctx, types.KeyConfirmationHeight, &h)
	return h
}

func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

func (k Keeper) SetKeyIDByAddress(ctx sdk.Context, address string, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keyIDbyAddrPrefix+address), []byte(keyID))
}

func (k Keeper) GetKeyIDByAddress(ctx sdk.Context, address string) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keyIDbyAddrPrefix + address))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) SetKeyIDByOutpoint(ctx sdk.Context, outpoint *wire.OutPoint, keyID string) {
	ctx.KVStore(k.storeKey).Set([]byte(keyIDbyOutPointPrefix+outpoint.String()), []byte(keyID))
}

func (k Keeper) GetKeyIDByOutpoint(ctx sdk.Context, outpoint *wire.OutPoint) (string, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(keyIDbyOutPointPrefix + outpoint.String()))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) GetRawTx(ctx sdk.Context, outpoint *wire.OutPoint) *wire.MsgTx {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + outpoint.String()))
	if bz == nil {
		return nil
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx
}

func (k Keeper) SetRawTx(ctx sdk.Context, outpoint *wire.OutPoint, tx *wire.MsgTx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+outpoint.String()), bz)
}

func (k Keeper) HasVerifiedOutPoint(ctx sdk.Context, outPoint *wire.OutPoint) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(outPointPrefix + outPoint.String()))
}

func (k Keeper) GetVerifiedOutPointInfo(ctx sdk.Context, outPoint *wire.OutPoint) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(outPointPrefix + outPoint.String()))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var out types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &out)

	return out, true
}

func (k Keeper) SetUnverifiedOutpointInfo(ctx sdk.Context, info types.OutPointInfo) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(pendingPrefix+info.OutPoint.String()), bz)
}

func (k Keeper) GetUnverifiedOutPointInfo(ctx sdk.Context, outpoint *wire.OutPoint) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + outpoint.String()))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var info types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)

	return info, true
}

// ProcessVerificationResult stores the info related to the specified outpoint (format txID:voutIdx) permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessVerificationResult(ctx sdk.Context, outPoint string, verified bool) error {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + outPoint))
	if bz == nil {
		return fmt.Errorf("outpoint %s not found", outPoint)
	}
	if verified {
		ctx.KVStore(k.storeKey).Set([]byte(outPointPrefix+outPoint), bz)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pendingPrefix + outPoint))
	return nil
}

// GenerateDepositAddressAndRedeemScript creates a Bitcoin address to deposit tokens for a transfer to the recipient address,
// as well as the corresponding redeem script to spend it. The generated address is unique for each recipient.
func (k Keeper) GenerateDepositAddressAndRedeemScript(ctx sdk.Context, pk btcec.PublicKey, recipient balance.CrossChainAddress) (btcutil.Address, []byte, error) {
	redeemScript, err := createRedeemScript(pk, recipient)
	if err != nil {
		return nil, nil, err
	}
	hash := sha256.Sum256(redeemScript)
	addr, err := btcutil.NewAddressWitnessScriptHash(hash[:], k.getNetwork(ctx).Params)
	if err != nil {
		return nil, nil, err
	}
	return addr, redeemScript, nil
}

func (k Keeper) SetRedeemScript(ctx sdk.Context, address btcutil.Address, script []byte) {
	ctx.KVStore(k.storeKey).Set([]byte(scriptPrefix+address.String()), script)
}

func (k Keeper) GetRedeemScript(ctx sdk.Context, address btcutil.Address) ([]byte, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(scriptPrefix + address.String()))
	return bz, bz != nil
}

func (k Keeper) GetHashToSign(ctx sdk.Context, rawTx *wire.MsgTx) ([]byte, error) {
	if len(rawTx.TxIn) != 1 {
		return nil, fmt.Errorf("transaction must have exactly one input")
	}

	addr, err := k.getDepositAddress(ctx, &rawTx.TxIn[0].PreviousOutPoint)
	if err != nil {
		return nil, err
	}
	script, ok := k.GetRedeemScript(ctx, addr)
	if !ok {
		return nil, fmt.Errorf("could not find a redeem script for outpoint %s", rawTx.TxIn[0].PreviousOutPoint.String())
	}
	return txscript.CalcWitnessSigHash(script, txscript.NewTxSigHashes(rawTx), txscript.SigHashAll, rawTx, 0, rawTx.TxOut[0].Value)
}

func (k Keeper) AssembleBtcTx(ctx sdk.Context, rawTx *wire.MsgTx, sig btcec.Signature) (*wire.MsgTx, error) {
	addr, err := k.getDepositAddress(ctx, &rawTx.TxIn[0].PreviousOutPoint)
	if err != nil {
		return nil, err
	}

	witness, err := k.createWitness(ctx, sig, addr)
	if err != nil {
		return nil, err
	}
	rawTx.TxIn[0].Witness = witness

	payScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}

	if err := validateTxScripts(rawTx, payScript); err != nil {
		return nil, err
	}
	return rawTx, nil
}

func (k Keeper) getDepositAddress(ctx sdk.Context, outpoint *wire.OutPoint) (btcutil.Address, error) {
	out, ok := k.GetVerifiedOutPointInfo(ctx, outpoint)
	if !ok {
		return nil, fmt.Errorf("transaction ID is not known")
	}
	addr, err := btcutil.DecodeAddress(out.DepositAddr, k.getNetwork(ctx).Params)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (k Keeper) getNetwork(ctx sdk.Context) types.Network {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return network
}

func (k Keeper) createWitness(ctx sdk.Context, sig btcec.Signature, address btcutil.Address) (wire.TxWitness, error) {
	sigBytes := append(sig.Serialize(), byte(txscript.SigHashAll))
	redeemScript, ok := k.GetRedeemScript(ctx, address)
	if !ok {
		return nil, fmt.Errorf("redeem script for address %s not found", address.String())
	}
	return wire.TxWitness{sigBytes, redeemScript}, nil
}

func validateTxScripts(tx *wire.MsgTx, pkScript []byte) error {
	flags := txscript.StandardVerifyFlags

	// execute (dry-run) the public key and signature script to validate them
	scriptEngine, err := txscript.NewEngine(pkScript, tx, 0, flags, nil, nil, tx.TxOut[0].Value)
	if err != nil {
		return sdkerrors.Wrap(err, "could not create execution engine, aborting")
	}
	if err := scriptEngine.Execute(); err != nil {
		return sdkerrors.Wrap(err, "transaction failed to execute, aborting")
	}
	return nil
}

func createRedeemScript(pk btcec.PublicKey, crossAddr balance.CrossChainAddress) ([]byte, error) {
	keyBz := pk.SerializeCompressed()
	nonce := btcutil.Hash160([]byte(crossAddr.String()))

	redeemScript, err := txscript.NewScriptBuilder().AddData(keyBz).AddOp(txscript.OP_CHECKSIG).AddData(nonce).AddOp(txscript.OP_DROP).Script()
	if err != nil {
		return nil, err
	}
	return redeemScript, nil
}
