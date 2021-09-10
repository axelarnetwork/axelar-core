package tests

import (
	"fmt"
	mathRand "math/rand"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"

	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	ethereumAddress = "0xE3deF8C6b7E357bf38eC701Ce631f78F2532987A"
)

func TestOutPointInfo_Equals(t *testing.T) {
	// Take care to have identical slices with different pointers
	var bz1, bz2 []byte
	for _, b := range rand.I64GenBetween(0, 256).Take(chainhash.HashSize) {
		bz1 = append(bz1, byte(b))
		bz2 = append(bz2, byte(b))
	}
	hash1, err := chainhash.NewHash(bz1)
	if err != nil {
		panic(err)
	}
	hash2, err := chainhash.NewHash(bz2)
	if err != nil {
		panic(err)
	}

	op1 := types.OutPointInfo{
		OutPoint: wire.NewOutPoint(hash1, 3).String(),
		Amount:   0,
		Address:  "recipient",
	}

	op2 := types.OutPointInfo{
		OutPoint: wire.NewOutPoint(hash2, 3).String(),
		Amount:   0,
		Address:  "recipient",
	}

	assert.True(t, op1.Equals(op2))
	assert.Equal(t, op1, op2)
}

func TestNewMasterConsolidationAddress(t *testing.T) {
	repeat := 100

	internalPrivKey1, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	internalPrivKey2, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	internalPubKey1 := tss.Key{ID: rand.Str(10), Value: internalPrivKey1.PublicKey, Role: tss.MasterKey}
	internalPubKey2 := tss.Key{ID: rand.Str(10), Value: internalPrivKey2.PublicKey, Role: tss.MasterKey}

	externalKeyCount := 6
	externalKeyThreshold := 3

	var externalKeys []tss.Key
	var externalPrivKeys []*btcec.PrivateKey

	for i := 0; i < externalKeyCount; i++ {
		externalPrivKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}

		externalPrivKeys = append(externalPrivKeys, externalPrivKey)
		externalKeys = append(externalKeys, tss.Key{ID: rand.Str(10), Value: externalPrivKey.PublicKey, Role: tss.ExternalKey})
	}

	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}

	signWithExternalKeys := func(sigHash []byte) []btcec.Signature {
		var sigs []btcec.Signature

		for _, externalPrivKey := range externalPrivKeys {
			sig, err := externalPrivKey.Sign(sigHash)
			if err != nil {
				panic(err)
			}

			sigs = append(sigs, *sig)
		}

		return sigs
	}

	t.Run("should not be spendable by invalid combination of keys at any time", testutils.Func(func(t *testing.T) {
		internalKeyLockTime := time.Now()
		externalKeyLockTime := time.Now().AddDate(0, 0, int(rand.I64Between(1, 100)))

		address := types.NewMasterConsolidationAddress(internalPubKey1, internalPubKey2, int64(externalKeyThreshold), externalKeys, internalKeyLockTime, externalKeyLockTime, types.Testnet3)
		inputs := []types.OutPointToSign{
			{
				AddressInfo: address,
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount,
					address.Address,
				),
			},
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		types.AddOutput(tx, address.GetAddress(), outputAmount)
		tx.LockTime = uint32(externalKeyLockTime.AddDate(0, 0, int(rand.I64Between(-1000, 1000))).Unix())
		tx = types.EnableTimelock(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		internalSig1, err := internalPrivKey1.Sign(sigHash)
		assert.NoError(t, err)
		internalSig2, err := internalPrivKey2.Sign(sigHash)
		assert.NoError(t, err)
		externalSigs := signWithExternalKeys(sigHash)

		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*internalSig1, *internalSig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*internalSig2, *internalSig1}})
		assert.Error(t, err)

		for i := 0; i < externalKeyCount-externalKeyThreshold; i++ {
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{externalSigs[i:]})
			assert.Error(t, err)
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append([]btcec.Signature{*internalSig1}, externalSigs[i:]...)})
			assert.Error(t, err)
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append([]btcec.Signature{*internalSig2}, externalSigs[i:]...)})
			assert.Error(t, err)
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append(externalSigs[i:i+int(externalKeyThreshold)], *internalSig1)})
			assert.Error(t, err)
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append(externalSigs[i:i+int(externalKeyThreshold)], *internalSig2)})
			assert.Error(t, err)
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append([]btcec.Signature{*internalSig1, *internalSig2}, externalSigs[i:i+int(externalKeyThreshold)]...)})
			assert.Error(t, err)
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append([]btcec.Signature{*internalSig2, *internalSig1}, externalSigs[i:i+int(externalKeyThreshold)]...)})
			assert.Error(t, err)
		}
	}).Repeat(repeat))

	t.Run("should not be spendable by internal keys before the internal timelock elapses", testutils.Func(func(t *testing.T) {
		internalKeyLockTime := time.Now()
		externalKeyLockTime := time.Now().AddDate(0, 0, int(rand.I64Between(1, 100)))

		address := types.NewMasterConsolidationAddress(internalPubKey1, internalPubKey2, int64(externalKeyThreshold), externalKeys, internalKeyLockTime, externalKeyLockTime, types.Testnet3)
		inputs := []types.OutPointToSign{
			{
				AddressInfo: address,
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount,
					address.Address,
				),
			},
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		types.AddOutput(tx, address.GetAddress(), outputAmount)
		tx.LockTime = uint32(internalKeyLockTime.AddDate(0, 0, -int(rand.I64Between(1, 100))).Unix())
		tx = types.EnableTimelock(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		internalSig1, err := internalPrivKey1.Sign(sigHash)
		assert.NoError(t, err)
		internalSig2, err := internalPrivKey2.Sign(sigHash)
		assert.NoError(t, err)

		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*internalSig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*internalSig2}})
		assert.Error(t, err)
	}).Repeat(repeat))

	t.Run("should not be spendable by external keys before the external timelock elapses", testutils.Func(func(t *testing.T) {
		internalKeyLockTime := time.Now()
		externalKeyLockTime := time.Now().AddDate(0, 0, int(rand.I64Between(1, 100)))

		address := types.NewMasterConsolidationAddress(internalPubKey1, internalPubKey2, int64(externalKeyThreshold), externalKeys, internalKeyLockTime, externalKeyLockTime, types.Testnet3)
		inputs := []types.OutPointToSign{
			{
				AddressInfo: address,
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount,
					address.Address,
				),
			},
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		types.AddOutput(tx, address.GetAddress(), outputAmount)
		tx.LockTime = uint32(externalKeyLockTime.AddDate(0, 0, -int(rand.I64Between(1, 100))).Unix())
		tx = types.EnableTimelock(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		externalSigs := signWithExternalKeys(sigHash)

		for i := 0; i < externalKeyCount-externalKeyThreshold+1; i++ {
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{externalSigs[i : i+int(externalKeyThreshold)]})
			assert.Error(t, err)
		}
	}).Repeat(repeat))

	t.Run("should be spendable by internal keys and external keys anytime", testutils.Func(func(t *testing.T) {
		internalKeyLockTime := time.Now()
		externalKeyLockTime := time.Now().AddDate(0, 0, int(rand.I64Between(1, 100)))

		address := types.NewMasterConsolidationAddress(internalPubKey1, internalPubKey2, int64(externalKeyThreshold), externalKeys, internalKeyLockTime, externalKeyLockTime, types.Testnet3)
		inputs := []types.OutPointToSign{
			{
				AddressInfo: address,
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount,
					address.Address,
				),
			},
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		types.AddOutput(tx, address.GetAddress(), outputAmount)
		tx.LockTime = uint32(internalKeyLockTime.AddDate(0, 0, int(rand.I64Between(-1000, 1000))).Unix())
		tx = types.EnableTimelock(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		internalSig1, err := internalPrivKey1.Sign(sigHash)
		assert.NoError(t, err)
		internalSig2, err := internalPrivKey2.Sign(sigHash)
		assert.NoError(t, err)
		externalSigs := signWithExternalKeys(sigHash)

		for i := 0; i < externalKeyCount-externalKeyThreshold+1; i++ {
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append([]btcec.Signature{*internalSig1}, externalSigs[i:i+int(externalKeyThreshold)]...)})
			assert.NoError(t, err)
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{append([]btcec.Signature{*internalSig2}, externalSigs[i:i+int(externalKeyThreshold)]...)})
			assert.NoError(t, err)
		}
	}).Repeat(repeat))

	t.Run("should be spendable by internal keys after the internal timelock elapses", testutils.Func(func(t *testing.T) {
		internalKeyLockTime := time.Now()
		externalKeyLockTime := time.Now().AddDate(0, 0, int(rand.I64Between(1, 100)))

		address := types.NewMasterConsolidationAddress(internalPubKey1, internalPubKey2, int64(externalKeyThreshold), externalKeys, internalKeyLockTime, externalKeyLockTime, types.Testnet3)
		inputs := []types.OutPointToSign{
			{
				AddressInfo: address,
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount,
					address.Address,
				),
			},
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		types.AddOutput(tx, address.GetAddress(), outputAmount)
		tx.LockTime = uint32(internalKeyLockTime.AddDate(0, 0, int(rand.I64Between(1, 100))).Unix())
		tx = types.EnableTimelock(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		internalSig1, err := internalPrivKey1.Sign(sigHash)
		assert.NoError(t, err)
		internalSig2, err := internalPrivKey2.Sign(sigHash)
		assert.NoError(t, err)

		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*internalSig1}})
		assert.NoError(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*internalSig2}})
		assert.NoError(t, err)
	}).Repeat(repeat))

	t.Run("should be spendable by external keys after the external timelock elapses", testutils.Func(func(t *testing.T) {
		internalKeyLockTime := time.Now()
		externalKeyLockTime := time.Now().AddDate(0, 0, int(rand.I64Between(1, 100)))

		address := types.NewMasterConsolidationAddress(internalPubKey1, internalPubKey2, int64(externalKeyThreshold), externalKeys, internalKeyLockTime, externalKeyLockTime, types.Testnet3)
		inputs := []types.OutPointToSign{
			{
				AddressInfo: address,
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount,
					address.Address,
				),
			},
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		types.AddOutput(tx, address.GetAddress(), outputAmount)
		tx.LockTime = uint32(externalKeyLockTime.AddDate(0, 0, int(rand.I64Between(1, 100))).Unix())
		tx = types.EnableTimelock(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		externalSigs := signWithExternalKeys(sigHash)

		for i := 0; i < externalKeyCount-externalKeyThreshold+1; i++ {
			_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{externalSigs[i : i+int(externalKeyThreshold)]})
			assert.NoError(t, err)
		}
	}).Repeat(repeat))
}

func TestNewDepositAddress_SpendableByTheFirstKey(t *testing.T) {
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	pubKey1 := tss.Key{ID: rand.Str(10), Value: privKey1.PublicKey, Role: tss.MasterKey}
	pubKey2 := tss.Key{ID: rand.Str(10), Value: privKey2.PublicKey, Role: tss.SecondaryKey}

	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	linkedAddressInfo := types.NewDepositAddress(pubKey1, pubKey2, types.Testnet3, nexus.CrossChainAddress{Chain: evm.Ethereum, Address: ethereumAddress})
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}
	inputs := []types.OutPointToSign{
		{
			AddressInfo: linkedAddressInfo,
			OutPointInfo: types.NewOutPointInfo(
				outPoint,
				inputAmount, // 1btc
				linkedAddressInfo.Address,
			),
		},
	}

	tx := types.CreateTx()
	for _, input := range inputs {
		assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
	}
	types.AddOutput(tx, linkedAddressInfo.GetAddress(), outputAmount)

	sigHash, err := txscript.CalcWitnessSigHash(linkedAddressInfo.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
	assert.NoError(t, err)

	sig, err := privKey1.Sign(sigHash)
	assert.NoError(t, err)

	_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig}})
	assert.NoError(t, err)

	sig, err = privKey2.Sign(sigHash)
	assert.NoError(t, err)
	_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig}})
	assert.NoError(t, err)
}

func TestNewDepositAddress_SpendableByTheSecondKey(t *testing.T) {
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	pubKey1 := tss.Key{ID: rand.Str(10), Value: privKey1.PublicKey, Role: tss.MasterKey}
	pubKey2 := tss.Key{ID: rand.Str(10), Value: privKey2.PublicKey, Role: tss.SecondaryKey}

	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	linkedAddressInfo := types.NewDepositAddress(pubKey1, pubKey2, types.Testnet3, nexus.CrossChainAddress{Chain: evm.Ethereum, Address: ethereumAddress})
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}
	inputs := []types.OutPointToSign{
		{
			AddressInfo: linkedAddressInfo,
			OutPointInfo: types.NewOutPointInfo(
				outPoint,
				inputAmount, // 1btc
				linkedAddressInfo.Address,
			),
		},
	}

	tx := types.CreateTx()
	for _, input := range inputs {
		assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
	}
	types.AddOutput(tx, linkedAddressInfo.GetAddress(), outputAmount)

	sigHash, err := txscript.CalcWitnessSigHash(linkedAddressInfo.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
	assert.NoError(t, err)

	sig, err := privKey2.Sign(sigHash)
	assert.NoError(t, err)

	_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig}})
	assert.NoError(t, err)
}

func TestNewDepositAddress_NotSpendableByARandomKey(t *testing.T) {
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	randomPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	pubKey1 := tss.Key{ID: rand.Str(10), Value: privKey1.PublicKey, Role: tss.MasterKey}
	pubKey2 := tss.Key{ID: rand.Str(10), Value: privKey2.PublicKey, Role: tss.SecondaryKey}

	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	linkedAddressInfo := types.NewDepositAddress(pubKey1, pubKey2, types.Testnet3, nexus.CrossChainAddress{Chain: evm.Ethereum, Address: ethereumAddress})
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}
	inputs := []types.OutPointToSign{
		{
			AddressInfo: linkedAddressInfo,
			OutPointInfo: types.NewOutPointInfo(
				outPoint,
				inputAmount, // 1btc
				linkedAddressInfo.Address,
			),
		},
	}

	tx := types.CreateTx()
	for _, input := range inputs {
		assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
	}
	types.AddOutput(tx, linkedAddressInfo.GetAddress(), outputAmount)

	sigHash, err := txscript.CalcWitnessSigHash(linkedAddressInfo.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
	assert.NoError(t, err)

	sig, err := randomPrivateKey.Sign(sigHash)
	assert.NoError(t, err)

	_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig}})
	assert.Error(t, err)
}

func TestNewAnyoneCanSpendAddress(t *testing.T) {
	t.Run("should return an address that is spendable by anyone", testutils.Func(func(t *testing.T) {
		inputAmount := btcutil.Amount(100000000) // 1btc
		outputAmount := btcutil.Amount(10000000) // 0.1btc
		addressInfo := types.NewAnyoneCanSpendAddress(types.Testnet3)
		outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
		if err != nil {
			panic(err)
		}
		inputs := []types.OutPointToSign{
			{
				AddressInfo: types.AddressInfo{
					Address:      addressInfo.Address,
					RedeemScript: addressInfo.RedeemScript,
				},
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount, // 1btc
					addressInfo.Address,
				),
			},
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		types.AddOutput(tx, addressInfo.GetAddress(), outputAmount)

		tx.TxIn[0].Witness = wire.TxWitness{addressInfo.RedeemScript}

		payScript, err := txscript.PayToAddrScript(addressInfo.GetAddress())
		assert.NoError(t, err)

		scriptEngine, err := txscript.NewEngine(payScript, tx, 0, txscript.StandardVerifyFlags, nil, nil, int64(inputAmount))
		assert.NoError(t, err)

		err = scriptEngine.Execute()
		assert.NoError(t, err)
	}))
}

func TestEstimateTxSize(t *testing.T) {
	repeats := 100

	t.Run("should return a reasonable transaction size estimation", testutils.Func(func(t *testing.T) {
		privKey1, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		privKey2, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		pubKey1 := tss.Key{ID: rand.Str(10), Value: privKey1.PublicKey, Role: tss.MasterKey}
		pubKey2 := tss.Key{ID: rand.Str(10), Value: privKey2.PublicKey, Role: tss.SecondaryKey}

		inputCount := rand.I64Between(11, 20)
		outputCount := rand.I64Between(1, 11)
		var inputs []types.OutPointToSign

		for i := 0; i < int(inputCount); i++ {
			addressInfo := types.NewDepositAddress(pubKey1, pubKey2, types.Testnet3, nexus.CrossChainAddress{Chain: evm.Ethereum, Address: ethereumAddress})
			outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:%d", rand.HexStr(64), rand.I64Between(0, 100)))
			if err != nil {
				panic(err)
			}
			inputAmount := btcutil.Amount(rand.I64Between(100, 10000))

			inputs = append(inputs, types.OutPointToSign{
				AddressInfo: addressInfo,
				OutPointInfo: types.NewOutPointInfo(
					outPoint,
					inputAmount, // 1btc
					addressInfo.Address,
				),
			})
		}

		tx := types.CreateTx()
		for _, input := range inputs {
			assert.NoError(t, types.AddInput(tx, input.OutPointInfo.OutPoint))
		}
		for i := 0; i < int(outputCount); i++ {
			addressInfo := types.NewSecondaryConsolidationAddress(pubKey1, types.Testnet3)
			outputAmount := btcutil.Amount(rand.I64Between(1, 100))

			types.AddOutput(tx, addressInfo.GetAddress(), outputAmount)
		}

		var signatures [][]btcec.Signature

		for i, input := range inputs {
			sigHash, err := txscript.CalcWitnessSigHash(input.AddressInfo.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, i, int64(input.OutPointInfo.Amount))
			assert.NoError(t, err)

			signature, err := privKey1.Sign(sigHash)
			assert.NoError(t, err)
			signatures = append(signatures, []btcec.Signature{*signature})
		}

		signedTx, err := types.AssembleBtcTx(tx, inputs, signatures)
		assert.NoError(t, err)

		expected := mempool.GetTxVirtualSize(btcutil.NewTx(signedTx))
		actual := types.EstimateTxSize(*tx, inputs)

		// expected - 1 * inputCount <= actual <= expected because a bitcoin signature can either contain 71 or 72 bytes
		// https://transactionfee.info/charts/bitcoin-script-ecdsa-length/#:~:text=The%20ECDSA%20signatures%20used%20in,normally%20taking%20up%2032%20bytes
		assert.LessOrEqual(t, expected, actual)
		assert.LessOrEqual(t, actual-1*inputCount, expected)
	}).Repeat(repeats))
}

func TestConfirmOutpointRequest_GetOutPoint(t *testing.T) {
	t.Run("case insensitive", testutils.Func(func(t *testing.T) {
		hash, _ := chainhash.NewHash(rand.Bytes(chainhash.HashSize))
		outpoint := wire.NewOutPoint(hash, mathRand.Uint32())
		info := types.NewOutPointInfo(outpoint, btcutil.Amount(rand.PosI64()), rand.StrBetween(5, 100))
		req1 := types.NewConfirmOutpointRequest(rand.Bytes(sdk.AddrLen), info)
		req2 := types.NewConfirmOutpointRequest(req1.Sender, info)

		var runes []rune
		flipDistr := rand.Bools(0.5)

		for _, r := range req1.OutPointInfo.OutPoint {
			if unicode.IsLetter(r) && flipDistr.Next() {
				runes = append(runes, unicode.ToUpper(r))
			} else {
				runes = append(runes, r)
			}
		}

		req1.OutPointInfo.OutPoint = string(runes)
		assert.Equal(t, req1.OutPointInfo.GetOutPoint(), req2.OutPointInfo.GetOutPoint())
	}).Repeat(20))
}

func TestAddress(t *testing.T) {
	t.Run("case insensitive", testutils.Func(func(t *testing.T) {
		addr, err := btcutil.NewAddressWitnessScriptHash(rand.Bytes(32), types.Mainnet.Params())
		assert.NoError(t, err)

		addrStr1 := addr.EncodeAddress()
		addrStr2 := strings.ToUpper(addrStr1)

		addr1, err := btcutil.DecodeAddress(addrStr1, types.Mainnet.Params())
		assert.NoError(t, err)
		addr2, err := btcutil.DecodeAddress(addrStr2, types.Mainnet.Params())
		assert.NoError(t, err)
		assert.NotEqual(t, addrStr1, addrStr2)
		assert.Equal(t, addr1, addr2)
	}).Repeat(20))
}
