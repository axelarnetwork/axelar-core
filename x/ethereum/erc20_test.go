package ethereum

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

const (
	// Used to test ERC20 marshalling of invocations
	contractID       = "testSC"
	erc20Transfer    = "transfer(address,uint256)"
	erc20TransferSel = "0xa9059cbb"
	erc20Addr        = "0x337c67618968370907da31daef3020238d01c9de"
	erc20Val         = "10000000000000000000"
	erc20PaddedAddr  = "0x000000000000000000000000337c67618968370907da31daef3020238d01c9de"
	erc20PaddedVal   = "0x0000000000000000000000000000000000000000000000008ac7230489e80000"
	erc20length      = 68

	// This mnemonic must be used when creating a ganache workspace, with at least two addresses with enough balance
	mnemonic = "invest cloud minimum mirror keen razor husband desert engine actual flower shop"

	// Used when attempting to retrieve the receipt
	maxReceiptAttempts = 10
)

/*
This test is based in the following tutorial about ERC20 parameter serialization:

https://medium.com/swlh/understanding-data-payloads-in-ethereum-transactions-354dbe995371
https://medium.com/mycrypto/why-do-we-need-transaction-data-39c922930e92
*/
func TestERC20Marshal(t *testing.T) {

	// test function selector
	assert.Equal(t, erc20TransferSel, types.CalcSelector(erc20Transfer))

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

	for i := 0; i < iterations; i++ {

		nonce := rand.Uint64()
		amount := big.NewInt(rand.Int63())
		gasLimit := rand.Uint64()
		gasPrice := big.NewInt(rand.Int63())
		chainID := big.NewInt(rand.Int63())
		data := make([]byte, dataLength)
		rand.Read(data)

		privateKey, err := getPrivateKey("m/44'/60'/0'/0/0")
		assert.NoError(t, err)
		addr := crypto.PubkeyToAddress(privateKey.PublicKey)

		tx1 := ethTypes.NewTransaction(nonce, addr, amount, gasLimit, gasPrice, data)
		tx2 := ethTypes.NewTransaction(nonce, addr, amount, gasLimit, gasPrice, data)

		signer := ethTypes.NewEIP155Signer(chainID)

		signedTx1, err := ethTypes.SignTx(tx1, signer, privateKey)
		assert.NoError(t, err)

		V1, R1, S1 := signedTx1.RawSignatureValues()

		hash := signer.Hash(tx1)

		sig, err := types.ToEthSignature(exported.Signature{R: R1, S: S1}, hash, privateKey.PublicKey)
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

// This test deploys an ERC20 mintable contract and mints tokens for a predetermined wallet.
// It requires ganache to be executing and initialized with the `mnemonic` constant.
// If ganache is not running, the test is skipped
func TestGanache(t *testing.T) {

	client, _ := ethclient.Dial("http://127.0.0.1:7545")
	_, err := client.ChainID(context.Background())

	if err != nil {
		t.Logf("Ganache not running, skipping this test (error: %v)", err)
		t.SkipNow()
	}

	deployerKey, err := getPrivateKey("m/44'/60'/0'/0/0")
	assert.NoError(t, err)
	deployerAddr := crypto.PubkeyToAddress(deployerKey.PublicKey)

	contractAddr := testDeploy(t, client, deployerKey)

	toKey, err := getPrivateKey("m/44'/60'/0'/0/1")
	assert.NoError(t, err)

	toAddr := crypto.PubkeyToAddress(toKey.PublicKey)

	testMint(t, client, deployerAddr, contractAddr, toAddr, deployerKey)
}

// Deploys the smart contract available for these tests. It avoids deployment via the contract ABI
// in favor of creating a raw transaction for the same purpose.
func testDeploy(t *testing.T, client *ethclient.Client, privateKey *ecdsa.PrivateKey) common.Address {

	byteCode := common.FromHex(MymintableBin)

	networkID, err := client.NetworkID(context.Background())
	assert.NoError(t, err)
	signer := ethTypes.NewEIP155Signer(networkID)
	var gasLimit uint64 = 3000000
	tssSigner := &mock.SignerMock{GetCurrentMasterKeyFunc: func(sdk.Context, balance.Chain) (ecdsa.PublicKey, bool) {
		return privateKey.PublicKey, true
	}}

	params := types.DeployParams{
		ByteCode: byteCode,
		GasLimit: gasLimit,
	}

	query := keeper.NewQuerier(client, keeper.Keeper{}, tssSigner)
	txBz, err := query(sdk.Context{}, []string{keeper.CreateDeployTx}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(params)})
	assert.NoError(t, err)
	var tx *ethTypes.Transaction
	testutils.Codec().MustUnmarshalJSON(txBz, &tx)

	signedTx, err := ethTypes.SignTx(tx, signer, privateKey)
	assert.NoError(t, err)
	err = client.SendTransaction(context.Background(), signedTx)
	assert.NoError(t, err)

	hash := signedTx.Hash()

	var receipt *ethTypes.Receipt

	// Ganache might not be able to instantly generate the receipt,
	// so we prepare the test for this possibility and allow it to retry
	for i := 0; i < maxReceiptAttempts; i++ {

		t.Logf("Trying to fetch receipt for Tx 0x%x", hash.Bytes())
		time.Sleep(1 * time.Second)
		receipt, err = client.TransactionReceipt(context.Background(), hash)

		if err == nil {

			t.Logf("Contract address: %s\n", receipt.ContractAddress.Hex())

			return receipt.ContractAddress
		}

		t.Logf("Error getting receipt: %v\n", err)
	}

	t.FailNow()

	return common.Address{}
}

// Mint tokens associated to the contract used by these tests and associate them to the given wallet.
// It avoids invoking the mint function throught the ABI in favor of creating a raw transaction for the same purpose.
func testMint(t *testing.T, client *ethclient.Client, creatorAddr, contractAddr, toAddr common.Address, privateKey *ecdsa.PrivateKey) {

	instance, err := NewMymintable(contractAddr, client)

	assert.NoError(t, err)

	originalAmount, err := instance.BalanceOf(&bind.CallOpts{}, toAddr)

	assert.NoError(t, err)

	t.Logf("Original ammount: %d", originalAmount)

	decimals, err := instance.Decimals(&bind.CallOpts{})
	assert.NoError(t, err)

	t.Logf("Decimals: %d", decimals)

	decBig := big.NewInt(int64(decimals))
	amount := big.NewInt(10)
	amount.Mul(amount, decBig)
	t.Logf("Amount: %d", amount)

	var gasLimit uint64 = 3000000
	tssSigner := &mock.SignerMock{GetCurrentMasterKeyFunc: func(sdk.Context, balance.Chain) (ecdsa.PublicKey, bool) {
		return privateKey.PublicKey, true
	}}

	params := types.MintParams{
		GasLimit:   gasLimit,
		Amount:     sdk.NewIntFromBigInt(amount),
		Recipient:  toAddr,
		ContractID: contractID,
	}

	query := keeper.NewQuerier(client, keeper.Keeper{}, tssSigner)
	txBz, err := query(sdk.Context{}, []string{keeper.CreateMintTx}, abci.RequestQuery{Data: testutils.Codec().MustMarshalJSON(params)})
	assert.NoError(t, err)
	var tx *ethTypes.Transaction
	testutils.Codec().MustUnmarshalJSON(txBz, &tx)

	networkID, err := client.NetworkID(context.Background())
	assert.NoError(t, err)
	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)
	err = client.SendTransaction(context.Background(), signedTx)
	assert.NoError(t, err)

	newAmount, err := instance.BalanceOf(&bind.CallOpts{}, toAddr)

	assert.NoError(t, err)

	t.Logf("New Amount: %d", newAmount)

	expectedAmount := big.NewInt(0).Add(originalAmount, amount)

	assert.Equal(t, expectedAmount, newAmount)

}

func getPrivateKey(derivation string) (*ecdsa.PrivateKey, error) {

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	path := hdwallet.MustParseDerivationPath(derivation)
	account, err := wallet.Derive(path, false)

	if err != nil {
		return nil, err
	}

	privateKey, err := wallet.PrivateKey(account)

	if err != nil {
		return nil, err
	}

	return privateKey, nil

}
