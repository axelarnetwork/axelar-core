package keeper

import (
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

func (k Keeper) HasVerifiedOutPoint(ctx sdk.Context, txID string) bool {
	return ctx.KVStore(k.storeKey).Has([]byte(outPointPrefix + txID))
}

func (k Keeper) getVerifiedOutPoint(ctx sdk.Context, txID string) (types.OutPointInfo, bool) {
	bz := ctx.KVStore(k.storeKey).Get([]byte(outPointPrefix + txID))
	if bz == nil {
		return types.OutPointInfo{}, false
	}
	var out types.OutPointInfo
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &out)

	return out, true
}

func (k Keeper) SetUnverifiedOutpoint(ctx sdk.Context, txID string, info types.OutPointInfo) error {
	minHeight := k.getRequiredConfirmationHeight(ctx)
	if info.Confirmations < minHeight {
		return fmt.Errorf("not enough confirmations, expected at least %d, got %d", minHeight, info.Confirmations)
	}
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(info)
	ctx.KVStore(k.storeKey).Set([]byte(pendingPrefix+txID), bz)
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

func (k Keeper) getPkScript(ctx sdk.Context, txID string) ([]byte, error) {
	out, ok := k.getVerifiedOutPoint(ctx, txID)
	if !ok {
		return nil, fmt.Errorf("transaction ID is not known")
	}
	network := k.getNetwork(ctx)
	addr, err := btcutil.DecodeAddress(out.Recipient, network.Params())
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

func (k Keeper) AssembleBtcTx(ctx sdk.Context, txID string, pk btcec.PublicKey, sig btcec.Signature) (*wire.MsgTx, error) {
	rawTx := k.GetRawTx(ctx, txID)
	if rawTx == nil {
		return nil, fmt.Errorf("withdraw tx for ID %s has not been prepared yet", txID)
	}

	sigScript, err := createSigScript(sig, pk)
	if err != nil {
		return nil, err
	}
	rawTx.TxIn[0].SignatureScript = sigScript

	pkScript, err := k.getPkScript(ctx, txID)
	if err != nil {
		return nil, err
	}
	if err := validateTxScripts(rawTx, pkScript); err != nil {
		return nil, err
	}
	return rawTx, nil
}

func (k Keeper) GetHashToSign(ctx sdk.Context, txID string) ([]byte, error) {
	script, err := k.getPkScript(ctx, txID)
	if err != nil {
		return nil, err
	}
	tx := k.GetRawTx(ctx, txID)
	return txscript.CalcSignatureHash(script, txscript.SigHashAll, tx, 0)
}

// GetAddress creates a Bitcoin pubKey hash address from a public key.
// We use Pay2PKH for added security over Pay2PK as well as for the benefit of getting a parsed address from the response of
// getrawtransaction() on the Bitcoin rpc client
// If a cross chain address is specified, the hash address is created using a nonce calculated from the cross chain address
func (k Keeper) GetAddress(ctx sdk.Context, pk btcec.PublicKey, crossAddr balance.CrossChainAddress) (btcutil.Address, error) {
	addr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk.SerializeCompressed()), k.getNetwork(ctx).Params())

	if err := crossAddr.Validate(); err == nil {
		//TODO: calculate with the cross chain address
	}

	return addr, sdkerrors.Wrap(err, "could not convert the given public key into a bitcoin address")
}

func (k Keeper) CreateTx(ctx sdk.Context, utxoID string, satoshi sdk.Coin, recipient btcutil.Address) (*wire.MsgTx, error) {
	out, ok := k.getVerifiedOutPoint(ctx, utxoID)
	if !ok {
		return nil, fmt.Errorf("transaction ID is not known")
	}

	addrScript, err := txscript.PayToAddrScript(recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
	}

	/*
		Creating a Bitcoin transaction one step at a time:
			1. Create the transaction message
			2. Get the output of the deposit transaction and convert it into the transaction input
			3. Create a new output
		See https://blog.hlongvu.com/post/t0xx5dejn3-Understanding-btcd-Part-4-Create-and-Sign-a-Bitcoin-transaction-with-btcd
	*/
	tx := wire.NewMsgTx(wire.TxVersion)

	// The signature script will be set later and we have no witness
	txIn := wire.NewTxIn(out.OutPoint, nil, nil)
	tx.AddTxIn(txIn)
	txOut := wire.NewTxOut(satoshi.Amount.Int64(), addrScript)
	tx.AddTxOut(txOut)
	return tx, nil
}

func createSigScript(sig btcec.Signature, pk btcec.PublicKey) ([]byte, error) {
	sigBytes := append(sig.Serialize(), byte(txscript.SigHashAll))

	keyBytes := pk.SerializeCompressed()

	sigScript, err := txscript.NewScriptBuilder().AddData(sigBytes).AddData(keyBytes).Script()
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
