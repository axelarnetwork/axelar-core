package tests

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types"

	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/exported"
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

func TestNewLinkedAddress_SpendableByMasterKey(t *testing.T) {
	masterPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	secondaryPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	masterKey := tss.Key{ID: rand.Str(10), Value: masterPrivateKey.PublicKey, Role: tss.MasterKey}
	secondaryKey := tss.Key{ID: rand.Str(10), Value: secondaryPrivateKey.PublicKey, Role: tss.SecondaryKey}

	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	linkedAddress := types.NewLinkedAddress(masterKey, secondaryKey, types.Testnet3, nexus.CrossChainAddress{Chain: ethereum.Ethereum, Address: ethereumAddress})
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}
	inputs := []types.OutPointToSign{
		{
			AddressInfo: linkedAddress,
			OutPointInfo: types.NewOutPointInfo(
				outPoint,
				inputAmount, // 1btc
				linkedAddress.Address.EncodeAddress(),
			),
		},
	}
	outputs := []types.Output{
		{
			Amount:    outputAmount,
			Recipient: linkedAddress.Address,
		},
	}

	tx, err := types.CreateTx(inputs, outputs)
	assert.NoError(t, err)

	sigHash, err := txscript.CalcWitnessSigHash(linkedAddress.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
	assert.NoError(t, err)

	sig, err := masterPrivateKey.Sign(sigHash)
	assert.NoError(t, err)

	_, err = types.AssembleBtcTx(tx, inputs, []btcec.Signature{*sig})
	assert.NoError(t, err)
}

func TestNewLinkedAddress_SpendableBySecondaryKey(t *testing.T) {
	masterPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	secondaryPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	masterKey := tss.Key{ID: rand.Str(10), Value: masterPrivateKey.PublicKey, Role: tss.MasterKey}
	secondaryKey := tss.Key{ID: rand.Str(10), Value: secondaryPrivateKey.PublicKey, Role: tss.SecondaryKey}

	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	linkedAddress := types.NewLinkedAddress(masterKey, secondaryKey, types.Testnet3, nexus.CrossChainAddress{Chain: ethereum.Ethereum, Address: ethereumAddress})
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}
	inputs := []types.OutPointToSign{
		{
			AddressInfo: linkedAddress,
			OutPointInfo: types.NewOutPointInfo(
				outPoint,
				inputAmount, // 1btc
				linkedAddress.Address.EncodeAddress(),
			),
		},
	}
	outputs := []types.Output{
		{
			Amount:    outputAmount,
			Recipient: linkedAddress.Address,
		},
	}

	tx, err := types.CreateTx(inputs, outputs)
	assert.NoError(t, err)

	sigHash, err := txscript.CalcWitnessSigHash(linkedAddress.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
	assert.NoError(t, err)

	sig, err := secondaryPrivateKey.Sign(sigHash)
	assert.NoError(t, err)

	_, err = types.AssembleBtcTx(tx, inputs, []btcec.Signature{*sig})
	assert.NoError(t, err)
}

func TestNewLinkedAddress_NotSpendableByRandomKey(t *testing.T) {
	masterPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	secondaryPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	randomPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	masterKey := tss.Key{ID: rand.Str(10), Value: masterPrivateKey.PublicKey, Role: tss.MasterKey}
	secondaryKey := tss.Key{ID: rand.Str(10), Value: secondaryPrivateKey.PublicKey, Role: tss.SecondaryKey}

	inputAmount := btcutil.Amount(100000000) // 1btc
	outputAmount := btcutil.Amount(10000000) // 0.1btc
	linkedAddress := types.NewLinkedAddress(masterKey, secondaryKey, types.Testnet3, nexus.CrossChainAddress{Chain: ethereum.Ethereum, Address: ethereumAddress})
	outPoint, err := types.OutPointFromStr(fmt.Sprintf("%s:0", rand.HexStr(64)))
	if err != nil {
		panic(err)
	}
	inputs := []types.OutPointToSign{
		{
			AddressInfo: linkedAddress,
			OutPointInfo: types.NewOutPointInfo(
				outPoint,
				inputAmount, // 1btc
				linkedAddress.Address.EncodeAddress(),
			),
		},
	}
	outputs := []types.Output{
		{
			Amount:    outputAmount,
			Recipient: linkedAddress.Address,
		},
	}

	tx, err := types.CreateTx(inputs, outputs)
	assert.NoError(t, err)

	sigHash, err := txscript.CalcWitnessSigHash(linkedAddress.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, 0, int64(inputAmount))
	assert.NoError(t, err)

	sig, err := randomPrivateKey.Sign(sigHash)
	assert.NoError(t, err)

	_, err = types.AssembleBtcTx(tx, inputs, []btcec.Signature{*sig})
	assert.Error(t, err)
}

func TestEstimateTxSize(t *testing.T) {
	repeats := 100

	t.Run("should give reasonable transaction size estimation", testutils.Func(func(t *testing.T) {
		masterPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		secondaryPrivateKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		masterKey := tss.Key{ID: rand.Str(10), Value: masterPrivateKey.PublicKey, Role: tss.MasterKey}
		secondaryKey := tss.Key{ID: rand.Str(10), Value: secondaryPrivateKey.PublicKey, Role: tss.SecondaryKey}

		inputCount := rand.I64Between(11, 20)
		outputCount := rand.I64Between(1, 11)
		var inputs []types.OutPointToSign
		var outputs []types.Output

		for i := 0; i < int(inputCount); i++ {
			addressInfo := types.NewLinkedAddress(masterKey, secondaryKey, types.Testnet3, nexus.CrossChainAddress{Chain: ethereum.Ethereum, Address: ethereumAddress})
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
					addressInfo.Address.EncodeAddress(),
				),
			})
		}

		for i := 0; i < int(outputCount); i++ {
			addressInfo := types.NewConsolidationAddress(masterKey, types.Testnet3)
			outputAmount := btcutil.Amount(rand.I64Between(1, 100))

			outputs = append(outputs, types.Output{
				Amount:    outputAmount,
				Recipient: addressInfo.Address,
			})
		}

		tx, err := types.CreateTx(inputs, outputs)
		assert.NoError(t, err)

		var signatures []btcec.Signature

		for i, input := range inputs {
			sigHash, err := txscript.CalcWitnessSigHash(input.AddressInfo.RedeemScript, txscript.NewTxSigHashes(tx), txscript.SigHashAll, tx, i, int64(input.OutPointInfo.Amount))
			assert.NoError(t, err)

			signature, err := masterPrivateKey.Sign(sigHash)
			assert.NoError(t, err)
			signatures = append(signatures, *signature)
		}

		signedTx, err := types.AssembleBtcTx(tx, inputs, signatures)
		assert.NoError(t, err)

		expected := int64(signedTx.SerializeSize())
		actual := types.EstimateTxSize(*tx, inputs)

		assert.LessOrEqual(t, expected, actual)
		assert.LessOrEqual(t, actual-2*inputCount, expected)
	}).Repeat(repeats))
}
