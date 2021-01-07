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
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	contractID = "testSC"
	network    = types.Network(types.Rinkeby)
)

var poll vote.PollMeta

func TestInstallSC(t *testing.T) {

	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	handler := NewHandler(k, &ethMock.RPCClientMock{}, &ethMock.VoterMock{}, &ethMock.SignerMock{})
	binary := common.FromHex(MymintableBin)
	_, err := handler(ctx, types.NewMsgInstallSC(sdk.AccAddress("sender"), contractID, binary))

	assert.Nil(t, err)
	assert.Equal(t, binary, k.GetSmartContract(ctx, contractID))

}

func TestVerifyTx_Deploy(t *testing.T) {

	// setup
	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := getPrivateKey("m/44'/60'/0'/0/0")
	assert.NoError(t, err)

	tx := generateDeploy(common.FromHex(MymintableBin))

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	// contract is missing
	handler := NewHandler(k, rpc, voter, signer)
	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, signedTx.Hash(), contractID))
	assert.NoError(t, err)

	assertMocks(t, signedTx.Hash(), voter, 0, false)

	// wrong master key
	k = keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	k.SetSmartContract(ctx, contractID, signedTx.Data())

	altSigner := &ethMock.SignerMock{

		GetCurrentMasterKeyFunc: func(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool) {

			key, _ := ecdsa.GenerateKey(elliptic.P256(), cryptoRand.Reader)
			return key.PublicKey, true
		},
	}

	handler = NewHandler(k, rpc, voter, altSigner)
	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, signedTx.Hash(), contractID))
	assert.NoError(t, err)

	assertMocks(t, signedTx.Hash(), voter, 1, false)

	// wrong hash
	k = keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	k.SetSmartContract(ctx, contractID, signedTx.Data())

	handler = NewHandler(k, rpc, voter, signer)
	wrongHash := common.BytesToHash([]byte(testutils.RandString(256)))
	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, wrongHash, contractID))
	assert.NoError(t, err)

	assertMocks(t, wrongHash, voter, 2, false)

	// everything correct
	k = keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	k.SetSmartContract(ctx, contractID, signedTx.Data())

	handler = NewHandler(k, rpc, voter, signer)

	_, err = handler(ctx, types.NewMsgVerifyDeployTx(sdk.AccAddress("sender"), network, signedTx.Hash(), contractID))
	assert.NoError(t, err)

	assertMocks(t, signedTx.Hash(), voter, 3, true)

	actualTX, ok := k.GetTX(ctx, signedTx.Hash().String())
	assert.True(t, ok)

	assert.Equal(t, signedTx.Hash().Bytes(), actualTX.Hash.Bytes())
	assert.Equal(t, contractID, actualTX.ContractID)
	assert.Equal(t, network, actualTX.Network)

}

func TestVerifyTx_Mint(t *testing.T) {

	//setup
	cdc := testutils.Codec()
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"))
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	networkID := big.NewInt(0).SetBytes([]byte(network))
	txBlockNum := big.NewInt(rand.Int63())

	privateKey, err := getPrivateKey("m/44'/60'/0'/0/0")
	assert.NoError(t, err)

	toAddr := common.HexToAddress(erc20Addr)
	amount, ok := big.NewInt(0).SetString(erc20Val, 10)
	assert.True(t, ok)

	contractAddr := common.HexToAddress("0xE1D849ED321D6075B81e5F37E01163bE9485fd13")

	tx := generateMint(contractAddr, toAddr, amount)

	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkID), privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signedTx)

	rpc, signer, voter := getVerifyMocks(signedTx, networkID, txBlockNum, privateKey.PublicKey)

	// wrong master key
	altSigner := &ethMock.SignerMock{

		GetCurrentMasterKeyFunc: func(ctx sdk.Context, chain exported.Chain) (ecdsa.PublicKey, bool) {

			key, _ := ecdsa.GenerateKey(elliptic.P256(), cryptoRand.Reader)
			return key.PublicKey, true
		},
	}

	handler := NewHandler(k, rpc, voter, altSigner)
	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, signedTx.Hash(), toAddr, sdk.NewIntFromBigInt(amount)))
	assert.NoError(t, err)

	assertMocks(t, signedTx.Hash(), voter, 0, false)

	// wrong hash
	handler = NewHandler(k, rpc, voter, signer)
	wrongHash := common.BytesToHash([]byte(testutils.RandString(256)))

	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, wrongHash, toAddr, sdk.NewIntFromBigInt(amount)))
	assert.NoError(t, err)

	assertMocks(t, wrongHash, voter, 1, false)

	// everything correct
	handler = NewHandler(k, rpc, voter, signer)
	_, err = handler(ctx, types.NewMsgVerifyMintTx(sdk.AccAddress("sender"), network, signedTx.Hash(), toAddr, sdk.NewIntFromBigInt(amount)))
	assert.NoError(t, err)

	assertMocks(t, signedTx.Hash(), voter, 2, true)

	actualTX, ok := k.GetTX(ctx, signedTx.Hash().String())
	assert.True(t, ok)

	assert.Equal(t, signedTx.Hash().Bytes(), actualTX.Hash.Bytes())
	assert.Equal(t, network, actualTX.Network)
	assert.Equal(t, amount, actualTX.Amount.BigInt())

}

func generateDeploy(byteCode []byte) *ethTypes.Transaction {

	nonce := rand.Uint64()
	gasPrice := big.NewInt(rand.Int63())
	gasLimit := rand.Uint64()
	value := big.NewInt(0)

	return ethTypes.NewContractCreation(nonce, value, gasLimit, gasPrice, byteCode)

}

func generateMint(contractAddr, toAddr common.Address, amount *big.Int) *ethTypes.Transaction {

	nonce := rand.Uint64()
	gasPrice := big.NewInt(rand.Int63())
	gasLimit := rand.Uint64()
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

func assertMocks(t *testing.T, hash common.Hash, voter *ethMock.VoterMock, index int, result bool) {

	assert.Equal(t, index+1, len(voter.InitPollCalls()))
	assert.Equal(t, hash.String(), voter.InitPollCalls()[index].Poll.ID)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), voter.InitPollCalls()[index].Poll.Type)
	assert.Equal(t, types.ModuleName, voter.InitPollCalls()[index].Poll.Module)

	assert.Equal(t, index+1, len(voter.RecordVoteCalls()))
	assert.Equal(t, poll, voter.RecordVoteCalls()[index].Vote.Poll())
	assert.Equal(t, result, voter.RecordVoteCalls()[index].Vote.Data())

}
