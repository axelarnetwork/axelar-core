package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
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
	Mainnet  = Network{Name: "main"}
	Testnet3 = Network{Name: "test"}
	Regtest  = Network{Name: "regtest"}
)

const (
	main    = "main"
	test    = "test"
	regtest = "regtest"
)

// maxDerSigLength defines the maximum size in bytes of a DER encoded bitcoin signature, and a bitcoin signature can only get up to 72 bytes according to
// https://transactionfee.info/charts/bitcoin-script-ecdsa-length/#:~:text=The%20ECDSA%20signatures%20used%20in,normally%20taking%20up%2032%20bytes
const maxDerSigLength = 72

// MinRelayTxFeeSatoshiPerByte defines bitcoin's default minimum relay fee in satoshi/byte
const MinRelayTxFeeSatoshiPerByte = int64(mempool.DefaultMinRelayTxFee / 1000)

// Params returns the network parameters
func (m Network) Params() *chaincfg.Params {
	switch m.Name {
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
func (m *Network) Validate() error {
	switch m.Name {
	case main, test, regtest:
		return nil
	default:
		return fmt.Errorf("unknown network: %s", m)
	}
}

// OutPointState is an enum for the state of an outpoint
type OutPointState int

// States of confirmed out points
const (
	CONFIRMED OutPointState = iota
	SPENT
)

// NewOutPointInfo returns a new OutPointInfo instance
func NewOutPointInfo(outPoint *wire.OutPoint, amount btcutil.Amount, address string) OutPointInfo {
	return OutPointInfo{
		OutPoint: outPoint.String(),
		Amount:   amount,
		Address:  address,
	}
}

// Validate ensures that all fields are filled with sensible values
func (m OutPointInfo) Validate() error {
	if _, err := OutPointFromStr(m.OutPoint); err != nil {
		return sdkerrors.Wrap(err, "outpoint malformed")
	}
	if m.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if m.Address == "" {
		return fmt.Errorf("invalid address to track")
	}
	return nil
}

// Equals checks if two OutPointInfo objects are semantically equal
func (m OutPointInfo) Equals(other OutPointInfo) bool {
	return m.OutPoint == other.OutPoint &&
		m.Amount == other.Amount &&
		m.Address == other.Address
}

func (m OutPointInfo) String() string {
	return m.OutPoint + "_" + m.Address + "_" + m.Amount.String()
}

// GetOutPoint returns the outpoint as a struct instead of a string
func (m OutPointInfo) GetOutPoint() wire.OutPoint {
	return *MustConvertOutPointFromStr(m.OutPoint)
}

// CreateTx returns a new unsigned Bitcoin transaction
func CreateTx(prevOuts []OutPointToSign, outputs []Output) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	for _, in := range prevOuts {
		outPoint, err := OutPointFromStr(in.OutPoint)
		if err != nil {
			return nil, err
		}
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

// MustConvertOutPointFromStr returns the parsed outpoint from a string of the form "txID:voutIdx". Panics if the string is malformed
func MustConvertOutPointFromStr(outStr string) *wire.OutPoint {
	o, err := OutPointFromStr(outStr)
	if err != nil {
		panic(err)
	}
	return o
}

// Output represents a Bitcoin transaction output
type Output struct {
	Amount    btcutil.Amount
	Recipient btcutil.Address
}

// RedeemScript represents the script that is used to redeem a transaction that spent to the address derived from the script
type RedeemScript []byte

// createCrossChainRedeemScript generates a redeem script unique to the given keys and cross-chain address
func createCrossChainRedeemScript(pk1 btcec.PublicKey, pk2 btcec.PublicKey, crossAddr nexus.CrossChainAddress) RedeemScript {
	nonce := btcutil.Hash160([]byte(crossAddr.String()))

	// the UTXOs sent to deposit addresses can be spent by both the master and secondary keys
	// therefore the redeem script requires a 1-of-2 multisig
	redeemScript, err := txscript.NewScriptBuilder().
		// Push a zero onto the stack and then swap it with the signature due to a bug in OP_CHECKMULTISIG that pops a dummy argument in the end and ignores it.
		// For more details, check out https://bitcoin.stackexchange.com/questions/40669/checkmultisig-a-worked-out-example/40673#40673
		AddOp(txscript.OP_0).
		AddOp(txscript.OP_SWAP).
		AddOp(txscript.OP_1).
		AddData(pk1.SerializeCompressed()).
		AddData(pk2.SerializeCompressed()).
		AddOp(txscript.OP_2).
		AddOp(txscript.OP_CHECKMULTISIG).
		AddData(nonce).
		AddOp(txscript.OP_DROP).
		Script()
	// The script builder only returns an error if the script is non-canonical.
	// Since we want to build canonical scripts and the template is predefined, an error here means the template is wrong,
	// i.e. it's a bug.
	if err != nil {
		panic(err)
	}
	return redeemScript
}

// createAnyoneCanSpendRedeemScript generates a redeem script that anyone can spend
func createAnyoneCanSpendRedeemScript() RedeemScript {
	redeemScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_TRUE).
		Script()
	// The script builder only returns an error if the script is non-canonical.
	// Since we want to build canonical scripts and the template is predefined, an error here means the template is wrong,
	// i.e. it's a bug.
	if err != nil {
		panic(err)
	}
	return redeemScript
}

// CreateMasterRedeemScript generates a redeem script unique to the given key
func CreateMasterRedeemScript(pk btcec.PublicKey) RedeemScript {
	redeemScript, err := txscript.NewScriptBuilder().
		AddData(pk.SerializeCompressed()).
		AddOp(txscript.OP_CHECKSIG).
		Script()
	// The script builder only returns an error if the script is non-canonical.
	// Since we want to build canonical scripts and the template is predefined, an error here means the template is wrong,
	// i.e. it's a bug.
	if err != nil {
		panic(err)
	}
	return redeemScript
}

// CreateP2WSHAddress creates a SeqWit script address based on a redeem script
func CreateP2WSHAddress(script RedeemScript, network Network) *btcutil.AddressWitnessScriptHash {
	hash := sha256.Sum256(script)
	// hash is 32 bit long, so this cannot throw an error if there is no bug
	addr, err := btcutil.NewAddressWitnessScriptHash(hash[:], network.Params())
	if err != nil {
		panic(err)
	}
	return addr
}

// NewConsolidationAddress creates a new address used to consolidate all unspent outpoints
func NewConsolidationAddress(pk tss.Key, network Network) AddressInfo {
	script := CreateMasterRedeemScript(btcec.PublicKey(pk.Value))
	address := CreateP2WSHAddress(script, network)

	return AddressInfo{
		RedeemScript: script,
		Address:      address.EncodeAddress(),
		Role:         Consolidation,
		KeyID:        pk.ID,
	}
}

// NewLinkedAddress creates a new address to make a deposit which can be transfered to another blockchain
func NewLinkedAddress(masterKey tss.Key, secondaryKey tss.Key, network Network, recipient nexus.CrossChainAddress) AddressInfo {
	script := createCrossChainRedeemScript(
		btcec.PublicKey(masterKey.Value),
		btcec.PublicKey(secondaryKey.Value),
		recipient,
	)
	address := CreateP2WSHAddress(script, network)

	return AddressInfo{
		RedeemScript: script,
		Address:      address.EncodeAddress(),
		Role:         Deposit,
		KeyID:        secondaryKey.ID,
	}
}

// NewAnyoneCanSpendAddress creates a p2wsh address that anyone can spend
func NewAnyoneCanSpendAddress(network Network) AddressInfo {
	script := createAnyoneCanSpendRedeemScript()
	address := CreateP2WSHAddress(script, network)

	return AddressInfo{
		RedeemScript: script,
		Address:      address.EncodeAddress(),
		Role:         None,
	}
}

// GetAddress returns the encoded bitcoin address
func (m AddressInfo) GetAddress() btcutil.Address {
	address, err := btcutil.DecodeAddress(m.Address, nil)
	if err != nil {
		panic(fmt.Errorf("invalid bitcoin address %s found", m.Address))
	}

	return address
}

// ToCrossChainAddr returns the corresponding cross-chain address
func (m AddressInfo) ToCrossChainAddr() nexus.CrossChainAddress {
	return nexus.CrossChainAddress{
		Chain:   exported.Bitcoin,
		Address: m.Address,
	}
}

// validateTxScript checks if the input at the given index can be spent with the given script
func validateTxScript(tx *wire.MsgTx, idx int, amount int64, payScript []byte) error {
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
	AddressInfo
}

// AssembleBtcTx assembles the unsigned transaction and given signature.
// Returns an error if the resulting signed Bitcoin transaction is invalid.
func AssembleBtcTx(rawTx *wire.MsgTx, outpointsToSign []OutPointToSign, sigs []btcec.Signature) (*wire.MsgTx, error) {
	for i, in := range outpointsToSign {
		sigBytes := append(sigs[i].Serialize(), byte(txscript.SigHashAll))
		rawTx.TxIn[i].Witness = wire.TxWitness{sigBytes, in.RedeemScript}

		payScript, err := txscript.PayToAddrScript(in.AddressInfo.GetAddress())
		if err != nil {
			return nil, err
		}

		if err := validateTxScript(rawTx, i, int64(in.OutPointInfo.Amount), payScript); err != nil {
			return nil, err
		}
	}

	return rawTx, nil
}

// MustEncodeTx serializes a given bitcoin transaction; panic if error
func MustEncodeTx(tx *wire.MsgTx) []byte {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// MustDecodeTx deserializes a bytes to a bitcoin transaction; panic if error
func MustDecodeTx(bz []byte) wire.MsgTx {
	var tx wire.MsgTx

	rbuf := bytes.NewReader(bz)
	if err := tx.Deserialize(rbuf); err != nil {
		panic(err)
	}

	return tx
}

// EstimateTxSize calculates the upper limit of the size in byte of given transaction after all witness data is attached
func EstimateTxSize(tx wire.MsgTx, outpointsToSign []OutPointToSign) int64 {
	for i, input := range outpointsToSign {
		zeroSigBytes := make([]byte, maxDerSigLength)
		tx.TxIn[i].Witness = wire.TxWitness{zeroSigBytes, input.RedeemScript}
	}

	return mempool.GetTxVirtualSize(btcutil.NewTx(&tx))
}

// Native asset denominations
const (
	Sat     = "sat"
	Satoshi = "satoshi"
	Btc     = "btc"
	Bitcoin = "bitcoin"
)

// ParseSatoshi parses a string to Satoshi, returning errors if invalid. Inputs in Bitcoin are automatically converted.
// This returns an error on an empty string as well.
func ParseSatoshi(rawCoin string) (sdk.Coin, error) {
	coin, err := sdk.ParseDecCoin(rawCoin)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("could not parse coin string")
	}

	switch coin.Denom {
	case Sat, Satoshi:
		break
	case Btc, Bitcoin:
		coin = sdk.NewDecCoinFromDec(Satoshi, coin.Amount.MulInt64(btcutil.SatoshiPerBitcoin))
	default:
		return sdk.Coin{}, fmt.Errorf("choose a correct denomination: %s (%s), %s (%s)", Satoshi, Sat, Bitcoin, Btc)
	}

	sat, remainder := coin.TruncateDecimal()
	if !remainder.IsZero() {
		return sdk.Coin{}, fmt.Errorf("amount in satoshi must be an integer value")
	}
	return sat, nil
}
