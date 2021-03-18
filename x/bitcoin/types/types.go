package types

import (
	"crypto/ecdsa"
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
)

// Bitcoin network types
var (
	Mainnet  = Network{&chaincfg.MainNetParams}
	Testnet3 = Network{&chaincfg.TestNet3Params}
	Regtest  = Network{&chaincfg.RegressionNetParams}
)

// OutPointInfo describes all the necessary information to verify the outPoint of a transaction
type OutPointInfo struct {
	OutPoint      *wire.OutPoint
	Amount        btcutil.Amount
	Address       string
	Confirmations int64
}

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
		OutPoint:      outPoint,
		Amount:        amount,
		Address:       txOut.ScriptPubKey.Addresses[0],
		Confirmations: txOut.Confirmations,
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

// Network provides additional functionality based on the bitcoin network name
type Network struct {
	// Params returns the configuration parameters associated with the chain
	Params *chaincfg.Params
}

// NetworkFromStr returns network given string
func NetworkFromStr(net string) (Network, error) {
	switch net {
	case "main":
		return Mainnet, nil
	case "test":
		return Testnet3, nil
	case "regtest":
		return Regtest, nil
	default:
		return Network{}, fmt.Errorf("unknown network: %s", net)
	}
}

// Validate checks if the object is a valid chain
func (n Network) Validate() error {
	if n.Params == nil {
		return fmt.Errorf("network could not be parsed, choose main, test, or regtest")
	}
	return nil
}

// RawTxParams describe the parameters used to create a raw unsigned transaction for Bitcoin
type RawTxParams struct {
	OutPoint    *wire.OutPoint
	DepositAddr string
	Satoshi     sdk.Coin
}

// CreateTx returns a new unsigned Bitcoin transaction
func CreateTx(prevOuts []*wire.OutPoint, outputs []Output) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	for _, outPoint := range prevOuts {
		// The signature script or witness will be set later
		txIn := wire.NewTxIn(outPoint, nil, nil)
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
func CreateDepositAddress(network Network, script RedeemScript) *btcutil.AddressWitnessScriptHash {
	hash := sha256.Sum256(script)
	// hash is 32 bit long, so this cannot throw an error if there is no bug
	addr, err := btcutil.NewAddressWitnessScriptHash(hash[:], network.Params)
	if err != nil {
		panic(err)
	}
	return addr
}

// ScriptAddress is a wrapper containing both Bitcoin P2WSH address and corresponding script
type ScriptAddress struct {
	*btcutil.AddressWitnessScriptHash
	RedeemScript RedeemScript
}

// NewConsolidationAddress creates a new address used to consolidate all unspent outpoints
func NewConsolidationAddress(pk ecdsa.PublicKey, network Network) ScriptAddress {
	script := CreateMasterRedeemScript(btcec.PublicKey(pk))
	addr := CreateDepositAddress(network, script)
	return ScriptAddress{
		RedeemScript:             script,
		AddressWitnessScriptHash: addr,
	}
}

// NewLinkedAddress creates a new address to make a deposit which can be transfered to another blockchain
func NewLinkedAddress(pk ecdsa.PublicKey, network Network, recipient nexus.CrossChainAddress) ScriptAddress {
	script := CreateCrossChainRedeemScript(btcec.PublicKey(pk), recipient)
	addr := CreateDepositAddress(network, script)
	return ScriptAddress{
		RedeemScript:             script,
		AddressWitnessScriptHash: addr,
	}
}

// ToCrossChainAddr returns the corresponding cross-chain address
func (addr ScriptAddress) ToCrossChainAddr() nexus.CrossChainAddress {
	return nexus.CrossChainAddress{
		Chain:   exported.Bitcoin,
		Address: addr.EncodeAddress(),
	}
}

// CreateTxWitness creates a transaction witness
func CreateTxWitness(sig btcec.Signature, redeemScript RedeemScript) wire.TxWitness {
	sigBytes := append(sig.Serialize(), byte(txscript.SigHashAll))
	return wire.TxWitness{sigBytes, redeemScript}
}

// ValidateTxScript checks if the input at the given index can be spent with the given script
func ValidateTxScript(tx *wire.MsgTx, idx int, input OutPointInfo, payScript []byte) error {
	// make sure the tx is considered standard to increase its chance to be mined
	flags := txscript.StandardVerifyFlags

	// execute (dry-run) the public key and signature script to validate them
	scriptEngine, err := txscript.NewEngine(payScript, tx, idx, flags, nil, nil, int64(input.Amount))
	if err != nil {
		return sdkerrors.Wrap(err, "could not create execution engine, aborting")
	}
	if err := scriptEngine.Execute(); err != nil {
		return sdkerrors.Wrap(err, "transaction failed to execute, aborting")
	}
	return nil
}
