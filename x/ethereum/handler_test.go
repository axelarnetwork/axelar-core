package ethereum

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptoRand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/balance/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	ethMock "github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	contractID   = "testSC"
	contractAddr = "0xE1D849ED321D6075B81e5F37E01163bE9485fd13"
	network      = types.Network(types.Rinkeby)
)

var poll vote.PollMeta

func TestInstallSC(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	handler := NewHandler(k, &ethMock.RPCClientMock{}, &ethMock.VoterMock{}, &ethMock.SignerMock{}, &ethMock.BalancerMock{})
	binary := common.FromHex(MymintableBin)
	_, err := handler(ctx, types.NewMsgInstallSC(sdk.AccAddress("sender"), contractID, binary))

	assert.NoError(t, err)
	assert.Equal(t, binary, k.GetSmartContract(ctx, contractID))

}
func TestVerifyTx_Deploy_ContractMissing(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	tx := generateDeploy(common.FromHex(MymintableBin))

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, signedTx.Hash(), contractID))
	assert.NoError(t, err)

	assertVoteCompleted(t, signedTx.Hash(), voter, false)

}

func TestVerifyTx_Deploy_WrongMK(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	tx := generateDeploy(common.FromHex(MymintableBin))

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)
	// wrong master key
	k = keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	k.SetSmartContract(ctx, contractID, signedTx.Data())

	signer.GetCurrentMasterKeyFunc = func(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool) {

		key, _ := ecdsa.GenerateKey(elliptic.P256(), cryptoRand.Reader)
		return key.PublicKey, true
	}

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, signedTx.Hash(), contractID))
	assert.NoError(t, err)

	assertVoteCompleted(t, signedTx.Hash(), voter, false)
}

func TestVerifyTx_Deploy_WrongTXHash(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	tx := generateDeploy(common.FromHex(MymintableBin))

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	k = keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	k.SetSmartContract(ctx, contractID, signedTx.Data())

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	wrongHash := common.BytesToHash([]byte(testutils.RandString(256)))
	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, wrongHash, contractID))
	assert.NoError(t, err)

	assertVoteCompleted(t, wrongHash, voter, false)

}

func TestVerifyTx_Deploy_Success(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	tx := generateDeploy(common.FromHex(MymintableBin))

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	k = keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	k.SetSmartContract(ctx, contractID, signedTx.Data())

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})

	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, signedTx.Hash(), contractID))
	assert.NoError(t, err)

	assertVoteCompleted(t, signedTx.Hash(), voter, true)

	actualTX, ok := k.GetTxForPoll(ctx, poll.ID)
	assert.True(t, ok)

	assert.Equal(t, signedTx.Hash().Bytes(), actualTX.Hash.Bytes())
	assert.Equal(t, contractID, actualTX.ContractID)
	assert.Equal(t, network, actualTX.Network)

}
func TestVerifyTx_Mint_WrongMK(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	toAddr := common.HexToAddress(erc20Addr)
	amount, ok := big.NewInt(0).SetString(erc20Val, 10)
	assert.True(t, ok)

	contractAddr := common.HexToAddress(contractAddr)

	tx := generateMint(contractAddr, toAddr, amount)

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signedTx)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	signer.GetCurrentMasterKeyFunc = func(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool) {

		key, _ := ecdsa.GenerateKey(elliptic.P256(), cryptoRand.Reader)
		return key.PublicKey, true

	}

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, signedTx.Hash(), toAddr, sdk.NewIntFromBigInt(amount)))
	assert.NoError(t, err)

	assertVoteCompleted(t, signedTx.Hash(), voter, false)
}
func TestVerifyTx_Mint_WrongTXHash(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	toAddr := common.HexToAddress(erc20Addr)
	amount, ok := big.NewInt(0).SetString(erc20Val, 10)
	assert.True(t, ok)

	contractAddr := common.HexToAddress(contractAddr)

	tx := generateMint(contractAddr, toAddr, amount)

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signedTx)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	wrongHash := common.BytesToHash([]byte(testutils.RandString(256)))

	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, wrongHash, toAddr, sdk.NewIntFromBigInt(amount)))
	assert.NoError(t, err)

	assertVoteCompleted(t, wrongHash, voter, false)
}

func TestVerifyTx_Mint_WrongToAddr(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	toAddr := common.HexToAddress(erc20Addr)
	amount, ok := big.NewInt(0).SetString(erc20Val, 10)
	assert.True(t, ok)

	contractAddr := common.HexToAddress(contractAddr)

	tx := generateMint(contractAddr, toAddr, amount)

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signedTx)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	wrongToAddr := common.BytesToAddress([]byte(testutils.RandString(256)))

	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, signedTx.Hash(), wrongToAddr, sdk.NewIntFromBigInt(amount)))
	assert.NoError(t, err)

	assertVoteCompleted(t, signedTx.Hash(), voter, false)
}

func TestVerifyTx_Mint_WrongAmount(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	toAddr := common.HexToAddress(erc20Addr)
	amount, ok := big.NewInt(0).SetString(erc20Val, 10)
	assert.True(t, ok)

	contractAddr := common.HexToAddress(contractAddr)

	tx := generateMint(contractAddr, toAddr, amount)

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signedTx)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	wrongAmount := big.NewInt(testutils.RandInts().Next())

	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, signedTx.Hash(), toAddr, sdk.NewIntFromBigInt(wrongAmount)))
	assert.NoError(t, err)

	assertVoteCompleted(t, signedTx.Hash(), voter, false)
}

func TestVerifyTx_Mint_Success(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := ethCrypto.GenerateKey()
	assert.NoError(t, err)

	toAddr := common.HexToAddress(erc20Addr)
	amount, ok := big.NewInt(0).SetString(erc20Val, 10)
	assert.True(t, ok)

	contractAddr := common.HexToAddress(contractAddr)

	tx := generateMint(contractAddr, toAddr, amount)

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signedTx)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	handler := NewHandler(k, rpc, voter, signer, &ethMock.BalancerMock{})
	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, signedTx.Hash(), toAddr, sdk.NewIntFromBigInt(amount)))
	assert.NoError(t, err)

	assertVoteCompleted(t, signedTx.Hash(), voter, true)

	actualTX, ok := k.GetTxForPoll(ctx, poll.ID)
	assert.True(t, ok)

	assert.Equal(t, signedTx.Hash().Bytes(), actualTX.Hash.Bytes())
	assert.Equal(t, network, actualTX.Network)
	assert.Equal(t, amount, actualTX.Amount.BigInt())

}

func generateDeploy(byteCode []byte) *ethTypes.Transaction {

	generator := testutils.RandInts()

	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(generator.Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)

	return ethTypes.NewContractCreation(nonce, value, gasLimit, gasPrice, byteCode)

}

func generateMint(contractAddr, toAddr common.Address, amount *big.Int) *ethTypes.Transaction {

	generator := testutils.RandInts()

	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(testutils.RandInts().Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)
	data := createMintCallData(toAddr, amount)

	return ethTypes.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data)

}

func getVerifyMocks(signedTx *ethTypes.Transaction, networkID, blockNum *big.Int, pk ecdsa.PublicKey) (*ethMock.RPCClientMock, *ethMock.SignerMock, *ethMock.VoterMock) {

	rpc := ethMock.RPCClientMock{

		NetworkIDFunc: func(ctx context.Context) (*big.Int, error) {
			return networkID, nil
		},

		TransactionByHashFunc: func(ctx context.Context, hash common.Hash) (tx *ethTypes.Transaction, isPending bool, err error) {

			if bytes.Equal(signedTx.Hash().Bytes(), hash.Bytes()) {

				return signedTx, false, nil
			}

			return nil, false, fmt.Errorf("transaction not found")
		},

		TransactionReceiptFunc: func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {

			if bytes.Equal(signedTx.Hash().Bytes(), hash.Bytes()) {

				return &ethTypes.Receipt{
					BlockNumber: blockNum,
				}, nil
			}

			return nil, fmt.Errorf("transaction not found")
		},

		BlockNumberFunc: func(ctx context.Context) (uint64, error) {

			lastBlockNum := big.NewInt(rand.Int63() + 1)
			return lastBlockNum.Add(lastBlockNum, blockNum).Uint64(), nil

		},
	}

	signer := ethMock.SignerMock{

		GetCurrentMasterKeyFunc: func(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool) {

			return pk, true
		},
	}

	voter := ethMock.VoterMock{
		InitPollFunc: func(_ sdk.Context, p vote.PollMeta) error { poll = p; return nil },
		RecordVoteFunc: func(ctx sdk.Context, vote vote.MsgVote) error {
			return nil
		},
	}

	return &rpc, &signer, &voter
}

func assertVoteCompleted(t *testing.T, hash common.Hash, voter *ethMock.VoterMock, result bool) {

	assert.Equal(t, 1, len(voter.InitPollCalls()))
	assert.Equal(t, hash.String(), voter.InitPollCalls()[0].Poll.ID)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), voter.InitPollCalls()[0].Poll.Type)
	assert.Equal(t, types.ModuleName, voter.InitPollCalls()[0].Poll.Module)

	assert.Equal(t, 1, len(voter.RecordVoteCalls()))
	assert.Equal(t, poll, voter.RecordVoteCalls()[0].Vote.Poll())
	assert.Equal(t, result, voter.RecordVoteCalls()[0].Vote.Data())

}
