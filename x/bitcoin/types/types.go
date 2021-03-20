package types

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// Bitcoin network types
var (
	Mainnet  = Network{"main"}
	Testnet3 = Network{"test"}
	Regtest  = Network{"regtest"}
)

// Network provides additional functionality based on the bitcoin network name
type Network struct {
	Name string
}

const (
	main    = "main"
	test    = "test"
	regtest = "regtest"
)

// Params returns the network parameters
func (n Network) Params() *chaincfg.Params {
	switch n.Name {
	case main:
		return &chaincfg.MainNetParams
	case test:
		return &chaincfg.TestNet3Params
	case regtest:
		return &chaincfg.RegressionNetParams
	default:
		panic("invalid network")
	}
}

// NetworkFromStr returns network given string
func NetworkFromStr(networkName string) (Network, error) {
	switch networkName {
	case main:
		return Mainnet, nil
	case test:
		return Testnet3, nil
	case regtest:
		return Regtest, nil
	default:
		return Network{}, fmt.Errorf("unknown network: %s", networkName)
	}
}

// Validate validates the network type
func (n Network) Validate() error {
	switch n.Name {
	case main, test, regtest:
		return nil
	default:
		return fmt.Errorf("unknown network: %s", n)
	}
}

// OutPointInfo describes all the necessary information to verify the outPoint of a transaction
type OutPointInfo struct {
	OutPoint *wire.OutPoint
	Amount   btcutil.Amount
	Address  string
}

// OutPointState is an enum for the state of an outpoint
type OutPointState int

// States of confirmed out points
const (
	CONFIRMED OutPointState = iota
	SPENT
)

// NewOutPointInfo returns a new OutPointInfo instance
func NewOutPointInfo(outPoint *wire.OutPoint, txOut btcjson.GetTxOutResult) (OutPointInfo, error) {
	amount, err := btcutil.NewAmount(txOut.Value)
	if err != nil {
		return OutPointInfo{}, err
	}

	if len(txOut.ScriptPubKey.Addresses) != 1 {
		return OutPointInfo{}, fmt.Errorf("only txOuts with single spendable address allowed")
	}
	return OutPointInfo{
		OutPoint: outPoint,
		Amount:   amount,
		Address:  txOut.ScriptPubKey.Addresses[0],
	}, nil
}

// Validate ensures that all fields are filled with sensible values
func (i OutPointInfo) Validate() error {
	if i.OutPoint == nil {
		return fmt.Errorf("missing outpoint")
	}
	if i.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if i.Address == "" {
		return fmt.Errorf("invalid address to track")
	}
	return nil
}

// Equals checks if two OutPointInfo objects are semantically equal
func (i OutPointInfo) Equals(other OutPointInfo) bool {
	return i.OutPoint.Hash.IsEqual(&other.OutPoint.Hash) &&
		i.OutPoint.Index == other.OutPoint.Index &&
		i.Amount == other.Amount &&
		i.Address == other.Address
}

func (i OutPointInfo) String() string {
	return i.OutPoint.String() + "_" + i.Address + "_" + i.Amount.String()
}

// RawTxParams describe the parameters used to create a raw unsigned transaction for Bitcoin
type RawTxParams struct {
	OutPoint    *wire.OutPoint
	DepositAddr string
	Satoshi     sdk.Coin
}

// CreateTx returns a new unsigned Bitcoin transaction
func CreateTx(prevOuts []OutPointToSign, outputs []Output) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	for _, in := range prevOuts {
		// The signature script or witness will be set later
		txIn := wire.NewTxIn(in.OutPoint, nil, nil)
		tx.AddTxIn(txIn)
	}
	for _, out := range outputs {
		addrScript, err := txscript.PayToAddrScript(out.Recipient)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
		}
		txOut := wire.NewTxOut(int64(out.Amount), addrScript)
		tx.AddTxOut(txOut)
	}

	return tx, nil
}

// OutPointFromStr returns the parsed outpoint from a string of the form "txID:voutIdx"
func OutPointFromStr(outStr string) (*wire.OutPoint, error) {
	outParams := strings.Split(outStr, ":")
	if len(outParams) != 2 {
		return nil, fmt.Errorf("you must pass the outpoint identifier in the form of \"txID:voutIdx\"")
	}

	v, err := strconv.ParseUint(outParams[1], 10, 32)
	if err != nil {
		return nil, err
	}
	hash, err := chainhash.NewHashFromStr(outParams[0])
	if err != nil {
		return nil, err
	}

	out := wire.NewOutPoint(hash, uint32(v))
	return out, nil
}

// Output represents a Bitcoin transaction output
type Output struct {
	Amount    btcutil.Amount
	Recipient btcutil.Address
}

// DepositQueryParams describe the parameters used to query for a Bitcoin deposit address
type DepositQueryParams struct {
	Address string
	Chain   string
}

// RedeemScript represents the script that is used to redeem a transaction that spent to the address derived from the script
type RedeemScript []byte

// CreateCrossChainRedeemScript generates a redeem script unique to the given key and cross-chain address
func CreateCrossChainRedeemScript(pk btcec.PublicKey, crossAddr nexus.CrossChainAddress) RedeemScript {
	keyBz := pk.SerializeCompressed()
	nonce := btcutil.Hash160([]byte(crossAddr.String()))

	redeemScript, err := txscript.NewScriptBuilder().AddData(keyBz).AddOp(txscript.OP_CHECKSIG).AddData(nonce).AddOp(txscript.OP_DROP).Script()
	// the script builder only returns an error if the script is non-canonical.
	// Since we want to build canonical scripts and the template is predefined, an error here means the template is wrong,
	// i.e. it's a bug.
	if err != nil {
		panic(err)
	}
	return redeemScript
}

// CreateMasterRedeemScript generates a redeem script unique to the given key
func CreateMasterRedeemScript(pk btcec.PublicKey) RedeemScript {
	keyBz := pk.SerializeCompressed()

	redeemScript, err := txscript.NewScriptBuilder().AddData(keyBz).AddOp(txscript.OP_CHECKSIG).Script()
	// the script builder only returns an error of the script is non-canonical.
	// Since we want to build canonical scripts and the template is predefined, an error here means the template is wrong,
	// i.e. it's a bug.
	if err != nil {
		panic(err)
	}
	return redeemScript
}

// CreateDepositAddress creates a SeqWit script address based on a redeem script
func CreateDepositAddress(script RedeemScript, network Network) *btcutil.AddressWitnessScriptHash {
	hash := sha256.Sum256(script)
	// hash is 32 bit long, so this cannot throw an error if there is no bug
	addr, err := btcutil.NewAddressWitnessScriptHash(hash[:], network.Params())
	if err != nil {
		panic(err)
	}
	return addr
}

// ScriptAddress is a wrapper containing the Bitcoin P2WSH address, it's corresponding script and the underlying key
type ScriptAddress struct {
	*btcutil.AddressWitnessScriptHash
	RedeemScript RedeemScript
	Key          tss.Key
}

// NewConsolidationAddress creates a new address used to consolidate all unspent outpoints
func NewConsolidationAddress(pk tss.Key, network Network) ScriptAddress {
	script := CreateMasterRedeemScript(btcec.PublicKey(pk.Value))
	addr := CreateDepositAddress(script, network)
	return ScriptAddress{
		RedeemScript:             script,
		AddressWitnessScriptHash: addr,
		Key:                      pk,
	}
}

// NewLinkedAddress creates a new address to make a deposit which can be transfered to another blockchain
func NewLinkedAddress(pk tss.Key, network Network, recipient nexus.CrossChainAddress) ScriptAddress {
	script := CreateCrossChainRedeemScript(btcec.PublicKey(pk.Value), recipient)
	addr := CreateDepositAddress(script, network)
	return ScriptAddress{
		RedeemScript:             script,
		AddressWitnessScriptHash: addr,
		Key:                      pk,
	}
}

// ToCrossChainAddr returns the corresponding cross-chain address
func (addr ScriptAddress) ToCrossChainAddr() nexus.CrossChainAddress {
	return nexus.CrossChainAddress{
		Chain:   exported.Bitcoin,
		Address: addr.EncodeAddress(),
	}
}

// ValidateTxScript checks if the input at the given index can be spent with the given script
func ValidateTxScript(tx *wire.MsgTx, idx int, amount int64, payScript []byte) error {
	// make sure the tx is considered standard to increase its chance to be mined
	flags := txscript.StandardVerifyFlags

	// execute (dry-run) the public key and signature script to validate them
	scriptEngine, err := txscript.NewEngine(payScript, tx, idx, flags, nil, nil, amount)
	if err != nil {
		return sdkerrors.Wrap(err, "could not create execution engine, aborting")
	}
	if err := scriptEngine.Execute(); err != nil {
		return sdkerrors.Wrap(err, "transaction failed to execute, aborting")
	}
	return nil
}

// OutPointToSign gathers all information needed to sign an outpoint
type OutPointToSign struct {
	OutPointInfo
	ScriptAddress
}

// AssembleBtcTx assembles the unsigned transaction and given signature.
// Returns an error if the resulting signed Bitcoin transaction is invalid.
func AssembleBtcTx(rawTx *wire.MsgTx, outpointsToSign []OutPointToSign, sigs []btcec.Signature) (*wire.MsgTx, error) {
	for i, in := range outpointsToSign {

		sigBytes := append(sigs[i].Serialize(), byte(txscript.SigHashAll))
		rawTx.TxIn[i].Witness = wire.TxWitness{sigBytes, in.RedeemScript}

		payScript, err := txscript.PayToAddrScript(in.AddressWitnessScriptHash)
		if err != nil {
			return nil, err
		}

		if err := ValidateTxScript(rawTx, i, int64(in.OutPointInfo.Amount), payScript); err != nil {
			return nil, err
		}
	}

	return rawTx, nil
}
