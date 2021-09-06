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
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	externalPrivKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	pubKey1 := tss.Key{ID: rand.Str(10), Value: privKey1.PublicKey, Role: tss.MasterKey}
	pubKey2 := tss.Key{ID: rand.Str(10), Value: privKey2.PublicKey, Role: tss.MasterKey}
	externalPubKey := tss.Key{ID: rand.Str(10), Value: externalPrivKey.PublicKey, Role: tss.ExternalKey}
	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}

	t.Run("should not be spendable by (pubKey1 or pubKey2) before the timelock elapses", testutils.Func(func(t *testing.T) {
		address := types.NewMasterConsolidationAddress(pubKey1, pubKey2, 1, []tss.Key{externalPubKey}, time.Now(), types.Testnet3)
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
		outputs := []types.Output{
			{
				Amount:    outputAmount,
				Recipient: address.GetAddress(),
			},
		}

		tx, err := types.CreateTx(inputs, outputs)
		assert.NoError(t, err)
		tx.LockTime = uint32(time.Now().AddDate(0, 0, -1).Unix())
		tx = types.EnableTimelockAndRBF(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		sig1, err := privKey1.Sign(sigHash)
		assert.NoError(t, err)
		sig2, err := privKey2.Sign(sigHash)
		assert.NoError(t, err)
		sig3, err := externalPrivKey.Sign(sigHash)
		assert.NoError(t, err)

		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig3}})
		assert.Error(t, err)
	}))

	t.Run("should be spendable by (pubKey1 or pubKey2) after the timelock elapses", testutils.Func(func(t *testing.T) {
		address := types.NewMasterConsolidationAddress(pubKey1, pubKey2, 1, []tss.Key{externalPubKey}, time.Now(), types.Testnet3)
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
		outputs := []types.Output{
			{
				Amount:    outputAmount,
				Recipient: address.GetAddress(),
			},
		}

		tx, err := types.CreateTx(inputs, outputs)
		assert.NoError(t, err)
		tx.LockTime = uint32(time.Now().AddDate(0, 0, 1).Unix())
		tx = types.EnableTimelockAndRBF(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		sig1, err := privKey1.Sign(sigHash)
		assert.NoError(t, err)
		sig2, err := privKey2.Sign(sigHash)
		assert.NoError(t, err)
		sig3, err := externalPrivKey.Sign(sigHash)
		assert.NoError(t, err)

		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1}})
		assert.NoError(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2}})
		assert.NoError(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig3}})
		assert.Error(t, err)
	}))

	t.Run("should be spendable by ((pubKey1 or pubKey2) and externalPubKey) before the timelock elapses", testutils.Func(func(t *testing.T) {
		address := types.NewMasterConsolidationAddress(pubKey1, pubKey2, 1, []tss.Key{externalPubKey}, time.Now(), types.Testnet3)
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
		outputs := []types.Output{
			{
				Amount:    outputAmount,
				Recipient: address.GetAddress(),
			},
		}

		tx, err := types.CreateTx(inputs, outputs)
		assert.NoError(t, err)
		tx.LockTime = uint32(time.Now().AddDate(0, 0, -1).Unix())
		tx = types.EnableTimelockAndRBF(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		sig1, err := privKey1.Sign(sigHash)
		assert.NoError(t, err)
		sig2, err := privKey2.Sign(sigHash)
		assert.NoError(t, err)
		sig3, err := externalPrivKey.Sign(sigHash)
		assert.NoError(t, err)

		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1, *sig3}})
		assert.NoError(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2, *sig3}})
		assert.NoError(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig3, *sig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig3, *sig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1, *sig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2, *sig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1, *sig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2, *sig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig3, *sig3}})
		assert.Error(t, err)
	}))

	t.Run("should be spendable by ((pubKey1 or pubKey2) and multiple externalPubKeys) before the timelock elapses", testutils.Func(func(t *testing.T) {
		externalKeyCount := 6
		externalKeyThreshold := int64(3)

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

		address := types.NewMasterConsolidationAddress(pubKey1, pubKey2, externalKeyThreshold, externalKeys, time.Now(), types.Testnet3)
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
		outputs := []types.Output{
			{
				Amount:    outputAmount,
				Recipient: address.GetAddress(),
			},
		}

		tx, err := types.CreateTx(inputs, outputs)
		assert.NoError(t, err)
		tx.LockTime = uint32(time.Now().AddDate(0, 0, -1).Unix())
		tx = types.EnableTimelockAndRBF(tx)

		sigHash, err := txscript.CalcWitnessSigHash(address.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
		assert.NoError(t, err)

		sig1, err := privKey1.Sign(sigHash)
		assert.NoError(t, err)
		sig2, err := privKey2.Sign(sigHash)
		assert.NoError(t, err)
		externalSig1, err := externalPrivKeys[0].Sign(sigHash)
		assert.NoError(t, err)
		externalSig4, err := externalPrivKeys[3].Sign(sigHash)
		assert.NoError(t, err)
		externalSig6, err := externalPrivKeys[5].Sign(sigHash)
		assert.NoError(t, err)

		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1, *externalSig1, *externalSig4, *externalSig6}})
		assert.NoError(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2, *externalSig1, *externalSig4, *externalSig6}})
		assert.NoError(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*externalSig1, *externalSig4, *externalSig6, *sig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*externalSig1, *externalSig4, *externalSig6, *sig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1, *sig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2, *sig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig1, *sig1}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*sig2, *sig2}})
		assert.Error(t, err)
		_, err = types.AssembleBtcTx(tx, inputs, [][]btcec.Signature{{*externalSig1, *externalSig4, *externalSig6, *externalSig1}})
		assert.Error(t, err)
	}))
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
	outputs := []types.Output{
		{
			Amount:    outputAmount,
			Recipient: linkedAddressInfo.GetAddress(),
		},
	}

	tx, err := types.CreateTx(inputs, outputs)
	assert.NoError(t, err)

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
	outputs := []types.Output{
		{
			Amount:    outputAmount,
			Recipient: linkedAddressInfo.GetAddress(),
		},
	}

	tx, err := types.CreateTx(inputs, outputs)
	assert.NoError(t, err)

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
	outputs := []types.Output{
		{
			Amount:    outputAmount,
			Recipient: linkedAddressInfo.GetAddress(),
		},
	}

	tx, err := types.CreateTx(inputs, outputs)
	assert.NoError(t, err)

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
		outputs := []types.Output{
			{
				Amount:    outputAmount,
				Recipient: addressInfo.GetAddress(),
			},
		}

		tx, err := types.CreateTx(inputs, outputs)
		assert.NoError(t, err)

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
		var outputs []types.Output

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

		for i := 0; i < int(outputCount); i++ {
			addressInfo := types.NewSecondaryConsolidationAddress(pubKey1, types.Testnet3)
			outputAmount := btcutil.Amount(rand.I64Between(1, 100))

			outputs = append(outputs, types.Output{
				Amount:    outputAmount,
				Recipient: addressInfo.GetAddress(),
			})
		}

		tx, err := types.CreateTx(inputs, outputs)
		assert.NoError(t, err)

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
