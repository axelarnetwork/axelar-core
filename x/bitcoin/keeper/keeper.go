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
	rawPrefix      = "raw_"
	outPointPrefix = "out_"
	pendingPrefix  = "pend_"
	addrPrefix     = "addr_"
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

func (k Keeper) getRequiredConfirmationHeight(ctx sdk.Context) uint64 {
	var h uint64
	k.params.Get(ctx, types.KeyConfirmationHeight, &h)
	return h
}

func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

func (k Keeper) GetRawTx(ctx sdk.Context, txID string) *wire.MsgTx {
	bz := ctx.KVStore(k.storeKey).Get([]byte(rawPrefix + txID))
	if bz == nil {
		return nil
	}
	var tx *wire.MsgTx
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &tx)

	return tx
}

func (k Keeper) SetRawTx(ctx sdk.Context, txID string, tx *wire.MsgTx) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(tx)
	ctx.KVStore(k.storeKey).Set([]byte(rawPrefix+txID), bz)
}

func (k Keeper) GetVerifiedOutPoint(ctx sdk.Context, outpoint wire.OutPoint) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(outPointPrefix + outpoint.String()))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var out types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &out)

	return out, true
}

func (k Keeper) SetUnverifiedOutpoint(ctx sdk.Context, info types.OutPointInfo) error {
	minHeight := k.getRequiredConfirmationHeight(ctx)
	if info.Confirmations < minHeight {
		return fmt.Errorf("not enough confirmations, expected at least %d, got %d", minHeight, info.Confirmations)
	}
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(pendingPrefix+info.OutPoint.String()), bz)
	return nil
}

func (k Keeper) GetUnverifiedOutPoint(ctx sdk.Context, txID string) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + txID))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var info types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)

	return info, true
}

// ProcessVerificationResult stores the OutPointInfo related to the txID permanently if confirmed or discards the data otherwise
func (k Keeper) ProcessVerificationResult(ctx sdk.Context, txID string, verified bool) error {
	bz := ctx.KVStore(k.storeKey).Get([]byte(pendingPrefix + txID))
	if bz == nil {
		return fmt.Errorf("tx %s not found", txID)
	}
	if verified {
		ctx.KVStore(k.storeKey).Set([]byte(outPointPrefix+txID), bz)
	}
	ctx.KVStore(k.storeKey).Delete([]byte(pendingPrefix + txID))
	return nil
}

func (k Keeper) AssembleBtcTx(ctx sdk.Context, rawTx *wire.MsgTx, pk btcec.PublicKey, sig btcec.Signature, recipient balance.CrossChainAddress) (*wire.MsgTx, error) {

	sigScript, err := createSigScript(sig, pk, recipient)
	if err != nil {
		return nil, err
	}
	rawTx.TxIn[0].SignatureScript = sigScript

	pkScript, err := k.getPayToAddrScript(ctx, rawTx.TxIn[0].PreviousOutPoint)
	if err != nil {
		return nil, err
	}
	if err := validateTxScripts(rawTx, pkScript); err != nil {
		return nil, err
	}
	return rawTx, nil
}

func (k Keeper) GetHashToSign(ctx sdk.Context, rawTx *wire.MsgTx) ([]byte, error) {
	if len(rawTx.TxIn) != 1 {
		return nil, fmt.Errorf("transaction must have exactly one input")
	}
	script, err := k.getPayToAddrScript(ctx, rawTx.TxIn[0].PreviousOutPoint)
	if err != nil {
		return nil, err
	}
	return txscript.CalcSignatureHash(script, txscript.SigHashAll, rawTx, 0)
}

// GetDepositAddress creates a Bitcoin address to deposit tokens for a transfer to the recipient address.
// This address is unique for each recipient.
func (k Keeper) GetDepositAddress(ctx sdk.Context, pk btcec.PublicKey, recipient balance.CrossChainAddress) (btcutil.Address, error) {
	redeemScript, err := getRedeemScript(pk, recipient)
	if err != nil {
		return nil, err
	}
	addr, err := btcutil.NewAddressWitnessScriptHash(sha256.New().Sum(redeemScript), k.getNetwork(ctx).Params())
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (k Keeper) getPayToAddrScript(ctx sdk.Context, outPoint wire.OutPoint) ([]byte, error) {
	out, ok := k.GetVerifiedOutPoint(ctx, outPoint)
	if !ok {
		return nil, fmt.Errorf("transaction ID is not known")
	}
	network := k.getNetwork(ctx)
	addr, err := btcutil.DecodeAddress(out.DepositAddr, network.Params())
	if err != nil {
		return nil, err
	}
	script, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}
	return script, nil
}

func (k Keeper) getNetwork(ctx sdk.Context) types.Network {
	var network types.Network
	k.params.Get(ctx, types.KeyNetwork, &network)
	return network
}

func createSigScript(sig btcec.Signature, pk btcec.PublicKey, recipient balance.CrossChainAddress) ([]byte, error) {
	sigBytes := append(sig.Serialize(), byte(txscript.SigHashAll))
	redeemScript, err := getRedeemScript(pk, recipient)
	if err != nil {
		return nil, err
	}
	sigScript, err := txscript.NewScriptBuilder().AddData(sigBytes).AddData(redeemScript).Script()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create bitcoin signature script")
	}
	return sigScript, nil
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

func getRedeemScript(pk btcec.PublicKey, crossAddr balance.CrossChainAddress) ([]byte, error) {
	keyBz := pk.SerializeCompressed()
	nonce := btcutil.Hash160([]byte(crossAddr.String()))

	redeemScript, err := txscript.NewScriptBuilder().AddData(keyBz).AddOp(txscript.OP_CHECKSIG).AddData(nonce).AddOp(txscript.OP_DROP).Script()
	if err != nil {
		return nil, err
	}
	return redeemScript, nil
}
