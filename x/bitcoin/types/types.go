package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

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

// CreateTx creates a new tx
func CreateTx() *wire.MsgTx {
	return wire.NewMsgTx(wire.TxVersion)
}

// AddInput adds the given input to the given tx
func AddInput(tx *wire.MsgTx, outPointStr string) error {
	outPoint, err := OutPointFromStr(outPointStr)
	if err != nil {
		return err
	}

	// The signature script or witness will be set later at signing time
	tx.AddTxIn(wire.NewTxIn(outPoint, nil, nil))

	return nil
}

// AddOutput adds the given address and amount as a new output to the given tx
func AddOutput(tx *wire.MsgTx, address btcutil.Address, amount btcutil.Amount) error {
	addrScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		return sdkerrors.Wrap(err, "could not create pay-to-address script for destination address")
	}

	tx.AddTxOut(wire.NewTxOut(int64(amount), addrScript))

	return nil
}

// GetOutputsTotal returns the total amount of outputs in the given tx
func GetOutputsTotal(tx wire.MsgTx) btcutil.Amount {
	total := btcutil.Amount(0)

	for _, output := range tx.TxOut {
		total += btcutil.Amount(output.Value)
	}

	return total
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

// createMultisigScript creates a 1-of-2 multisig script with a nonce providing uniqueness
func createMultisigScript(pubKey1 btcec.PublicKey, pubKey2 btcec.PublicKey, nonce []byte) RedeemScript {
	// the UTXOs sent to deposit addresses can be spent by both the master and secondary keys
	// therefore the redeem script requires a 1-of-2 multisig
	redeemScript, err := txscript.NewScriptBuilder().
		// Push a zero onto the stack and then swap it with the signature due to a bug in OP_CHECKMULTISIG that pops a dummy argument in the end and ignores it.
		// For more details, check out https://bitcoin.stackexchange.com/questions/40669/checkmultisig-a-worked-out-example/40673#40673
		AddOp(txscript.OP_0).
		AddOp(txscript.OP_SWAP).
		AddOp(txscript.OP_1).
		AddData(pubKey1.SerializeCompressed()).
		AddData(pubKey2.SerializeCompressed()).
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

// createP2pkScript generates a redeem script unique to the given key
func createP2pkScript(pubKey btcec.PublicKey) RedeemScript {
	redeemScript, err := txscript.NewScriptBuilder().
		AddData(pubKey.SerializeCompressed()).
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

func createMasterAddressScript(internalKey1 btcec.PublicKey, internalKey2 btcec.PublicKey, externalMultiSigThreshold int64, externalKeys []btcec.PublicKey, internalKeysOnlyLockTime time.Time, externalKeysOnlyLockTime time.Time) RedeemScript {
	if externalMultiSigThreshold <= 0 || externalMultiSigThreshold > int64(len(externalKeys)) {
		panic(fmt.Errorf("invalid external multisig threshold %d", externalMultiSigThreshold))
	}

	builder := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DEPTH).
		AddInt64(externalMultiSigThreshold + 1).
		AddOp(txscript.OP_EQUAL).
		AddOp(txscript.OP_IF)
	// if (externalMultiSigThreshold + 1) signatures exist on the stack
	for i := 0; i < int(externalMultiSigThreshold); i++ {
		builder = builder.AddOp(txscript.OP_TOALTSTACK)
	}

	builder = builder.AddOp(txscript.OP_0)
	for i := 0; i < int(externalMultiSigThreshold); i++ {
		builder = builder.AddOp(txscript.OP_FROMALTSTACK)
	}

	builder = builder.AddInt64(externalMultiSigThreshold)
	for _, externelKey := range externalKeys {
		builder = builder.AddData(externelKey.SerializeCompressed())
	}

	builder = builder.
		AddInt64(int64(len(externalKeys))).
		// Verify m/n multisig from external keys
		AddOp(txscript.OP_CHECKMULTISIGVERIFY).
		AddOp(txscript.OP_0).
		AddOp(txscript.OP_SWAP).
		AddInt64(1).
		AddData(internalKey1.SerializeCompressed()).
		AddData(internalKey2.SerializeCompressed()).
		AddInt64(2).
		// Verify 1/2 multisig from internalKey1 and internalKey2
		AddOp(txscript.OP_CHECKMULTISIG).
		AddOp(txscript.OP_ELSE).
		AddOp(txscript.OP_DEPTH).
		AddOp(txscript.OP_1).
		AddOp(txscript.OP_EQUAL).
		// if one signature exists on the stack
		AddOp(txscript.OP_IF).
		AddInt64(internalKeysOnlyLockTime.Unix()).
		AddOp(txscript.OP_CHECKLOCKTIMEVERIFY).
		// OP_DROP due to OP_CHECKLOCKTIMEVERIFY not popping anything from the stack
		AddOp(txscript.OP_DROP).
		AddOp(txscript.OP_0).
		AddOp(txscript.OP_SWAP).
		AddOp(txscript.OP_1).
		AddData(internalKey1.SerializeCompressed()).
		AddData(internalKey2.SerializeCompressed()).
		AddOp(txscript.OP_2).
		// Verify 1/2 multisig from internalKey1 and internalKey2
		AddOp(txscript.OP_CHECKMULTISIG).
		// if externalMultiSigThreshold signatures exist on the stack
		AddOp(txscript.OP_ELSE).
		AddOp(txscript.OP_DEPTH).
		AddInt64(externalMultiSigThreshold).
		AddOp(txscript.OP_EQUALVERIFY).
		AddInt64(externalKeysOnlyLockTime.Unix()).
		AddOp(txscript.OP_CHECKLOCKTIMEVERIFY).
		// OP_DROP due to OP_CHECKLOCKTIMEVERIFY not popping anything from the stack
		AddOp(txscript.OP_DROP)

	for i := 0; i < int(externalMultiSigThreshold); i++ {
		builder = builder.AddOp(txscript.OP_TOALTSTACK)
	}
	builder = builder.AddOp(txscript.OP_0)
	for i := 0; i < int(externalMultiSigThreshold); i++ {
		builder = builder.AddOp(txscript.OP_FROMALTSTACK)
	}
	builder = builder.AddInt64(externalMultiSigThreshold)
	for _, externelKey := range externalKeys {
		builder = builder.AddData(externelKey.SerializeCompressed())
	}
	builder = builder.
		AddInt64(int64(len(externalKeys))).
		// Verify m/n multisig from external keys
		AddOp(txscript.OP_CHECKMULTISIG)

	builder = builder.AddOp(txscript.OP_ENDIF).
		AddOp(txscript.OP_ENDIF)
	redeemScript, err := builder.Script()

	// The script builder only returns an error if the script is non-canonical.
	// Since we want to build canonical scripts and the template is predefined, an error here means the template is wrong,
	// i.e. it's a bug.
	if err != nil {
		panic(err)
	}
	return redeemScript
}

func createDepositAddressScript(pubKey btcec.PublicKey, externalMultiSigThreshold int64, externalKeys []btcec.PublicKey, externalKeysOnlyLockTime time.Time, nonce []byte) RedeemScript {
	builder := txscript.NewScriptBuilder().
		AddData(nonce).
		AddOp(txscript.OP_DROP).
		AddOp(txscript.OP_DEPTH).
		AddOp(txscript.OP_1).
		AddOp(txscript.OP_EQUAL).
		AddOp(txscript.OP_IF).
		AddData(pubKey.SerializeCompressed()).
		AddOp(txscript.OP_CHECKSIG).
		AddOp(txscript.OP_ELSE).
		AddOp(txscript.OP_DEPTH).
		AddInt64(externalMultiSigThreshold).
		AddOp(txscript.OP_EQUALVERIFY).
		AddInt64(externalKeysOnlyLockTime.Unix()).
		AddOp(txscript.OP_CHECKLOCKTIMEVERIFY).
		// OP_DROP due to OP_CHECKLOCKTIMEVERIFY not popping anything from the stack
		AddOp(txscript.OP_DROP)

	for i := 0; i < int(externalMultiSigThreshold); i++ {
		builder = builder.AddOp(txscript.OP_TOALTSTACK)
	}
	builder = builder.AddOp(txscript.OP_0)
	for i := 0; i < int(externalMultiSigThreshold); i++ {
		builder = builder.AddOp(txscript.OP_FROMALTSTACK)
	}
	builder = builder.AddInt64(externalMultiSigThreshold)
	for _, externelKey := range externalKeys {
		builder = builder.AddData(externelKey.SerializeCompressed())
	}
	builder = builder.
		AddInt64(int64(len(externalKeys))).
		// Verify m/n multisig from external keys
		AddOp(txscript.OP_CHECKMULTISIG)

	builder = builder.AddOp(txscript.OP_ENDIF)

	redeemScript, err := builder.Script()
	// The script builder only returns an error if the script is non-canonical.
	// Since we want to build canonical scripts and the template is predefined, an error here means the template is wrong,
	// i.e. it's a bug.
	if err != nil {
		panic(err)
	}
	return redeemScript
}

// createP2wshAddress creates a SeqWit script address based on a redeem script
func createP2wshAddress(script RedeemScript, network Network) *btcutil.AddressWitnessScriptHash {
	hash := sha256.Sum256(script)
	// hash is 32 bit long, so this cannot throw an error if there is no bug
	addr, err := btcutil.NewAddressWitnessScriptHash(hash[:], network.Params())
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMasterConsolidationAddress returns a p2wsh-wrapped address that is
// 1) spendable by the ((currMasterKey or oldMasterKey) and externalMultiSigThreshold/len(externalKeys) externalKeys) before the timelock elapses
// 2) spendable by the (currMasterKey or oldMasterKey) after the timelock elapses
func NewMasterConsolidationAddress(currMasterKey tss.Key, oldMasterKey tss.Key, externalMultiSigThreshold int64, externalKeys []tss.Key, internalKeysOnlyLockTime time.Time, externalKeysOnlyLockTime time.Time, network Network) AddressInfo {
	externalPubKeys := make([]btcec.PublicKey, len(externalKeys))
	for i, externalKey := range externalKeys {
		externalPubKeys[i] = btcec.PublicKey(externalKey.Value)
	}
	script := createMasterAddressScript(btcec.PublicKey(currMasterKey.Value), btcec.PublicKey(oldMasterKey.Value), externalMultiSigThreshold, externalPubKeys, internalKeysOnlyLockTime, externalKeysOnlyLockTime)
	address := createP2wshAddress(script, network)

	externalKeyIDs := make([]tss.KeyID, len(externalKeys))
	for i, externalKey := range externalKeys {
		externalKeyIDs[i] = externalKey.ID
	}

	return AddressInfo{
		RedeemScript: script,
		Address:      address.EncodeAddress(),
		Role:         Consolidation,
		KeyID:        currMasterKey.ID,
		MaxSigCount:  uint32(externalMultiSigThreshold) + 1,
		SpendingCondition: &AddressInfo_SpendingCondition{
			InternalKeyIds:            []tss.KeyID{currMasterKey.ID, oldMasterKey.ID},
			ExternalKeyIds:            externalKeyIDs,
			ExternalMultisigThreshold: externalMultiSigThreshold,
			LockTime:                  &internalKeysOnlyLockTime,
		},
	}
}

// NewSecondaryConsolidationAddress returns a p2wsh-wrapped p2pk address for the secondary key
func NewSecondaryConsolidationAddress(secondaryKey tss.Key, network Network) AddressInfo {
	script := createP2pkScript(btcec.PublicKey(secondaryKey.Value))
	address := createP2wshAddress(script, network)

	return AddressInfo{
		RedeemScript: script,
		Address:      address.EncodeAddress(),
		Role:         Consolidation,
		KeyID:        secondaryKey.ID,
		MaxSigCount:  1,
		SpendingCondition: &AddressInfo_SpendingCondition{
			InternalKeyIds:            []tss.KeyID{secondaryKey.ID},
			ExternalKeyIds:            []tss.KeyID{},
			ExternalMultisigThreshold: 0,
			LockTime:                  nil,
		},
	}
}

// NewDepositAddress returns a p2wsh-wrapped 1-of-2 multisig address that is spendable by the secondary or master key
// with a recipient cross chain address to provide uniqueness
func NewDepositAddress(secondaryKey tss.Key, externalMultiSigThreshold int64, externalKeys []tss.Key, externalKeysOnlyLockTime time.Time, recipient nexus.CrossChainAddress, network Network) AddressInfo {
	externalPubKeys := make([]btcec.PublicKey, len(externalKeys))
	for i, externalKey := range externalKeys {
		externalPubKeys[i] = btcec.PublicKey(externalKey.Value)
	}

	script := createDepositAddressScript(
		btcec.PublicKey(secondaryKey.Value),
		externalMultiSigThreshold,
		externalPubKeys,
		externalKeysOnlyLockTime,
		btcutil.Hash160([]byte(recipient.String())),
	)
	address := createP2wshAddress(script, network)

	return AddressInfo{
		RedeemScript: script,
		Address:      address.EncodeAddress(),
		Role:         Deposit,
		KeyID:        secondaryKey.ID,
		MaxSigCount:  1,
		SpendingCondition: &AddressInfo_SpendingCondition{
			InternalKeyIds:            []tss.KeyID{secondaryKey.ID},
			ExternalKeyIds:            []tss.KeyID{},
			ExternalMultisigThreshold: 0,
			LockTime:                  nil,
		},
	}
}

// NewAnyoneCanSpendAddress returns a p2wsh-wrapped anyone-can-spend address
func NewAnyoneCanSpendAddress(network Network) AddressInfo {
	script := createAnyoneCanSpendRedeemScript()
	address := createP2wshAddress(script, network)

	return AddressInfo{
		RedeemScript:      script,
		Address:           address.EncodeAddress(),
		Role:              None,
		KeyID:             "",
		MaxSigCount:       0,
		SpendingCondition: nil,
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
func AssembleBtcTx(rawTx *wire.MsgTx, outpointsToSign []OutPointToSign, sigs [][]btcec.Signature) (*wire.MsgTx, error) {
	for i, in := range outpointsToSign {
		witness := wire.TxWitness{}

		for _, sig := range sigs[i] {
			witness = append(witness, append(sig.Serialize(), byte(txscript.SigHashAll)))
		}
		rawTx.TxIn[i].Witness = append(witness, in.RedeemScript)

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

// MustDecodeAddress decodes the given address; panic if error
func MustDecodeAddress(address string, network Network) btcutil.Address {
	decoded, err := btcutil.DecodeAddress(address, network.Params())
	if err != nil {
		panic(err)
	}

	return decoded
}

// EstimateTxSize calculates the upper limit of the size in byte of given transaction after all witness data is attached
func EstimateTxSize(tx wire.MsgTx, outpointsToSign []OutPointToSign) int64 {
	zeroSigBytes := make([]byte, maxDerSigLength)

	for i, input := range outpointsToSign {
		var witness wire.TxWitness

		for j := 0; j < int(input.MaxSigCount); j++ {
			witness = append(witness, zeroSigBytes)
		}

		tx.TxIn[i].Witness = append(witness, input.RedeemScript)
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

// ToSatoshiCoin converts the given bitcoin or satoshi dec coin to the equivalent satoshi coin
func ToSatoshiCoin(coin sdk.DecCoin) (sdk.Coin, error) {
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

// ParseSatoshi parses a string to Satoshi, returning errors if invalid. Inputs in Bitcoin are automatically converted.
// This returns an error on an empty string as well.
func ParseSatoshi(rawCoin string) (sdk.Coin, error) {
	coin, err := sdk.ParseDecCoin(rawCoin)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("could not parse coin string")
	}

	return ToSatoshiCoin(coin)
}

// NewSignedTx is the constructor for SignedTx
func NewSignedTx(txType TxType, tx *wire.MsgTx, confirmationRequired bool, anyoneCanSpendVout uint32) SignedTx {
	return SignedTx{
		Type:                 txType,
		Tx:                   MustEncodeTx(tx),
		ConfirmationRequired: confirmationRequired,
		AnyoneCanSpendVout:   anyoneCanSpendVout,
	}
}

// GetTx gets the underlying tx
func (m SignedTx) GetTx() *wire.MsgTx {
	result := MustDecodeTx(m.Tx)
	return &result
}

// NewUnsignedTx is the constructor for UnsignedTx
func NewUnsignedTx(txType TxType, tx *wire.MsgTx, anyoneCanSpendVout uint32, internalTransferAmount btcutil.Amount) UnsignedTx {
	unsignedTx := UnsignedTx{
		Type:                   txType,
		Tx:                     MustEncodeTx(tx),
		Status:                 Created,
		ConfirmationRequired:   false,
		AnyoneCanSpendVout:     anyoneCanSpendVout,
		InternalTransferAmount: internalTransferAmount,
	}

	return unsignedTx
}

// SetTx sets the underlying tx
func (m *UnsignedTx) SetTx(tx *wire.MsgTx) {
	m.Tx = MustEncodeTx(tx)
}

// GetTx gets the underlying tx
func (m UnsignedTx) GetTx() *wire.MsgTx {
	result := MustDecodeTx(m.Tx)
	return &result
}

// Is returns true if unsigned transaction is in the given status; false otherwise
func (m UnsignedTx) Is(status TxStatus) bool {
	return m.Status == status
}

// DisableTimelock disables timelock(https://en.bitcoin.it/wiki/Timelock) on the given transaction.
func DisableTimelock(tx *wire.MsgTx) *wire.MsgTx {
	tx.LockTime = 0

	for i := range tx.TxIn {
		tx.TxIn[i].Sequence = wire.MaxTxInSequenceNum
	}

	return tx
}

// EnableTimelock enables timelock(https://en.bitcoin.it/wiki/Timelock) on the given transaction.
func EnableTimelock(tx *wire.MsgTx, lockTime uint32) *wire.MsgTx {
	tx.LockTime = lockTime

	for i := range tx.TxIn {
		tx.TxIn[i].Sequence = wire.MaxTxInSequenceNum - 1
	}

	return tx
}

// NewSigRequirement is the constructor for UnsignedTx_Info_InputInfo_SigRequirement
func NewSigRequirement(keyID tss.KeyID, sigHash []byte) UnsignedTx_Info_InputInfo_SigRequirement {
	return UnsignedTx_Info_InputInfo_SigRequirement{
		KeyID:   keyID,
		SigHash: sigHash,
	}
}

// GetTxTypes returns an array of all types of key role
func GetTxTypes() []TxType {
	var results []TxType

	for i := 1; i < len(TxType_value); i++ {
		results = append(results, TxType(i))
	}

	return results
}

// TxTypeFromSimpleStr creates a TxType from string
func TxTypeFromSimpleStr(str string) (TxType, error) {
	switch strings.ToLower(str) {
	case MasterConsolidation.SimpleString():
		return MasterConsolidation, nil
	case SecondaryConsolidation.SimpleString():
		return SecondaryConsolidation, nil
	case Rescue.SimpleString():
		return Rescue, nil
	default:
		return -1, fmt.Errorf("invalid tx type %s", str)
	}
}

// SimpleString returns a human-readable string
func (t TxType) SimpleString() string {
	switch t {
	case MasterConsolidation:
		return "master"
	case SecondaryConsolidation:
		return "secondary"
	case Rescue:
		return "rescue"
	default:
		return "unknown"
	}
}

// Validate validates the TxType
func (t TxType) Validate() error {
	txTypeStr, ok := TxType_name[int32(t)]
	if !ok || TxTypeUnspecified.String() == txTypeStr {
		return fmt.Errorf("invalid tx type %d", t)
	}

	return nil
}
