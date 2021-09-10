package keeper_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	// Used to test ERC20 marshalling of invocations
	erc20TransferSel = "0xa9059cbb"
	erc20Addr        = "0x337c67618968370907da31daef3020238d01c9de"
	erc20Val         = "10000000000000000000"
	erc20PaddedAddr  = "0x000000000000000000000000337c67618968370907da31daef3020238d01c9de"
	erc20PaddedVal   = "0x0000000000000000000000000000000000000000000000008ac7230489e80000"
	erc20length      = 68

	// This mnemonic creates deterministic wallet accounts
	mnemonic = "invest cloud minimum mirror keen razor husband desert engine actual flower shop"

	// DerivationPath describes the hierarchical deterministic wallet path to derive addresses
	DerivationPath = "m/44'/60'/0'/0/0"
)

/*
This test is based in the following tutorial about ERC20 parameter serialization:

https://medium.com/swlh/understanding-data-payloads-in-ethereum-transactions-354dbe995371
https://medium.com/mycrypto/why-do-we-need-transaction-data-39c922930e92
*/
func TestERC20Marshal(t *testing.T) {
	// test first parameter (the address)
	paddedAddr := hexutil.Encode(common.LeftPadBytes(common.HexToAddress(erc20Addr).Bytes(), 32))

	assert.Equal(t, erc20PaddedAddr, paddedAddr)

	// test second parameter (the amount)
	val, ok := big.NewInt(0).SetString(erc20Val, 10)
	assert.True(t, ok)

	paddedVal := hexutil.Encode(common.LeftPadBytes(val.Bytes(), 32))

	assert.Equal(t, erc20PaddedVal, paddedVal)

	// test total length of the data to be sent
	var data []byte

	data = append(data, common.FromHex(erc20TransferSel)...)
	data = append(data, common.FromHex(paddedAddr)...)
	data = append(data, common.FromHex(paddedVal)...)

	assert.Equal(t, erc20length, len(data))

}

const (
	dataLength = 256
	iterations = 32
)

func TestSig(t *testing.T) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		panic(err)
	}

	path := hdwallet.MustParseDerivationPath(DerivationPath)
	account, err := wallet.Derive(path, false)
	if err != nil {
		panic(err)
	}
	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		panic(err)
	}

	for i := 0; i < iterations; i++ {

		nonce := rand.Uint64()
		amount := big.NewInt(rand.Int63())
		gasLimit := rand.Uint64()
		gasPrice := big.NewInt(rand.Int63())
		chainID := big.NewInt(rand.Int63())
		data := make([]byte, dataLength)
		rand.Read(data)

		addr := crypto.PubkeyToAddress(privateKey.PublicKey)

		tx1 := evmTypes.NewTransaction(nonce, addr, amount, gasLimit, gasPrice, data)
		tx2 := evmTypes.NewTransaction(nonce, addr, amount, gasLimit, gasPrice, data)

		signer := evmTypes.NewEIP155Signer(chainID)

		signedTx1, err := evmTypes.SignTx(tx1, signer, privateKey)
		assert.NoError(t, err)

		V1, R1, S1 := signedTx1.RawSignatureValues()

		hash := signer.Hash(tx1)

		sig, err := types.ToSignature(tss.Signature{R: R1, S: S1}, hash, privateKey.PublicKey)
		assert.NoError(t, err)

		recoveredPK, err := crypto.SigToPub(hash.Bytes(), sig[:])
		assert.NoError(t, err)
		assert.Equal(t, privateKey.PublicKey.X.Bytes(), recoveredPK.X.Bytes())
		assert.Equal(t, privateKey.PublicKey.Y.Bytes(), recoveredPK.Y.Bytes())

		recoveredAddr := crypto.PubkeyToAddress(*recoveredPK)
		assert.Equal(t, addr, recoveredAddr)

		signedTx2, err := tx2.WithSignature(signer, sig[:])
		assert.NoError(t, err)

		V2, R2, S2 := signedTx2.RawSignatureValues()

		assert.Equal(t, V1, V2)
		assert.Equal(t, R1, R2)
		assert.Equal(t, S1, S2)

		expectedBZ := crypto.CompressPubkey(&privateKey.PublicKey)
		recoveredBZ := crypto.CompressPubkey(recoveredPK)

		assert.True(t, crypto.VerifySignature(expectedBZ, hash.Bytes(), sig[:64]))
		assert.True(t, crypto.VerifySignature(recoveredBZ, hash.Bytes(), sig[:64]))

		if sig[64] == 0 {
			sig[64] = 1
		} else {
			sig[64] = 0
		}

		recoveredPK, err = crypto.SigToPub(hash.Bytes(), sig[:])
		assert.NoError(t, err)

		recoveredAddr = crypto.PubkeyToAddress(*recoveredPK)
		assert.NotEqual(t, addr, recoveredAddr)
	}
}
