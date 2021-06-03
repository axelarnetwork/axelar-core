package keeper

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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

	// EthereumDerivationPath describes the hierarchical deterministic wallet path to derive Ethereum addresses
	EthereumDerivationPath = "m/44'/60'/0'/0/0"
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

	path := hdwallet.MustParseDerivationPath(EthereumDerivationPath)
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

		tx1 := ethTypes.NewTransaction(nonce, addr, amount, gasLimit, gasPrice, data)
		tx2 := ethTypes.NewTransaction(nonce, addr, amount, gasLimit, gasPrice, data)

		signer := ethTypes.NewEIP155Signer(chainID)

		signedTx1, err := ethTypes.SignTx(tx1, signer, privateKey)
		assert.NoError(t, err)

		V1, R1, S1 := signedTx1.RawSignatureValues()

		hash := signer.Hash(tx1)

		sig, err := types.ToEthSignature(tss.Signature{R: R1, S: S1}, hash, privateKey.PublicKey)
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

// Deploys the smart contract available for these tests. It avoids deployment via the contract ABI
// in favor of creating a raw transaction for the same purpose.
func TestDeploy(t *testing.T) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		panic(err)
	}

	path := hdwallet.MustParseDerivationPath(EthereumDerivationPath)
	account, err := wallet.Derive(path, false)
	if err != nil {
		panic(err)
	}
	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		panic(err)
	}

	addr, err := wallet.Address(account)
	if err != nil {
		panic(err)
	}

	backend := backends.NewSimulatedBackend(core.GenesisAlloc{addr: {Balance: big.NewInt(1 * params.Ether)}}, 3000000)

	chainID := backend.Blockchain().Config().ChainID
	assert.NoError(t, err)
	signer := ethTypes.NewEIP155Signer(chainID)
	var gasLimit uint64 = 3000000
	tssSigner := &mock.SignerMock{GetCurrentKeyFunc: func(_ sdk.Context, _ nexus.Chain, _ tss.KeyRole) (tss.Key, bool) {
		return tss.Key{
			ID:    rand2.StrBetween(5, 20),
			Value: privateKey.PublicKey,
			Role:  tss.MasterKey,
		}, true
	}}
	nexusMock := &mock.NexusMock{
		GetChainFunc: func(sdk.Context, string) (nexus.Chain, bool) {
			return nexus.Chain{
				Name:                  rand2.StrBetween(5, 20),
				NativeAsset:           rand2.StrBetween(3, 5),
				SupportsForeignAssets: true,
			}, true
		},
	}

	deployParams := types.DeployParams{
		GasPrice: sdk.ZeroInt(),
		GasLimit: gasLimit,
	}

	minConfHeight := rand2.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	encCfg := testutils.MakeEncodingConfig()

	rpc := &mock.RPCClientMock{PendingNonceAtFunc: backend.PendingNonceAt, SuggestGasPriceFunc: backend.SuggestGasPrice}
	query := NewQuerier(rpc, k, tssSigner, nexusMock)
	res, err := query(ctx, []string{CreateDeployTx}, abci.RequestQuery{Data: encCfg.Amino.MustMarshalJSON(deployParams)})
	assert.NoError(t, err)

	var result types.DeployResult
	encCfg.Amino.MustUnmarshalJSON(res, &result)

	signedTx, err := ethTypes.SignTx(result.Tx, signer, privateKey)
	assert.NoError(t, err)
	err = backend.SendTransaction(context.Background(), signedTx)
	assert.NoError(t, err)
	backend.Commit()

	t.Logf("Trying to fetch receipt for Tx %s", signedTx.Hash().String())
	contractAddr, err := bind.WaitDeployed(context.Background(), backend, signedTx)
	if err != nil {
		t.Logf("Error getting receipt: %v\n", err)
		t.FailNow()
	}
	t.Logf("Contract address: %s\n", contractAddr.Hex())
}
