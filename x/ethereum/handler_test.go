package ethereum

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	ethMock "github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	network = types.Network(types.Rinkeby)
)

var sender = sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 20))))

func TestVerifyTx_Deploy_HashNotFound(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedDeployTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return nil, false, fmt.Errorf("wrong hash")
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Deploy_WithoutData(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedDeployTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return ethTypes.NewContractCreation(signedTx.Nonce(), signedTx.Value(), signedTx.Gas(), signedTx.GasPrice(), nil), false, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Deploy_Pending(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	var confCount int64 = 0
	signedTx := createSignedDeployTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return signedTx, true, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Deploy_WrongValue(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedDeployTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return ethTypes.NewContractCreation(
			signedTx.Nonce(),
			big.NewInt(testutils.RandIntBetween(0, math.MaxInt64)),
			signedTx.Gas(),
			signedTx.GasPrice(),
			signedTx.Data(),
		), false, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Deploy_NotConfirmed(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(0, minConfHeight)
	signedTx := createSignedDeployTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return signedTx, false, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Deploy_Success(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedDeployTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), true)
}

func TestVerifyTx_Mint_HashNotFound(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedMintTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return nil, false, fmt.Errorf("wrong hash")
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_WithoutData(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedMintTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return ethTypes.NewContractCreation(signedTx.Nonce(), signedTx.Value(), signedTx.Gas(), signedTx.GasPrice(), nil), false, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_WrongRecipient(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedMintTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return ethTypes.NewContractCreation(signedTx.Nonce(), signedTx.Value(), signedTx.Gas(), signedTx.GasPrice(), nil), false, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_Pending(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	var confCount int64 = 0
	signedTx := createSignedMintTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return signedTx, true, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_WrongValue(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedMintTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return ethTypes.NewContractCreation(
			signedTx.Nonce(),
			big.NewInt(testutils.RandIntBetween(0, math.MaxInt64)),
			signedTx.Gas(),
			signedTx.GasPrice(),
			signedTx.Data(),
		), false, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_NotConfirmed(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(0, minConfHeight)
	signedTx := createSignedMintTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionByHashFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Transaction, bool, error) {
		return signedTx, false, nil
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_Success(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedMintTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, signedTx))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), true)
}

func createSignedDeployTx() *ethTypes.Transaction {
	generator := testutils.RandInts()

	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(generator.Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)
	byteCode := testutils.RandBytes(int(testutils.RandIntBetween(1, 10000)))

	return sign(ethTypes.NewContractCreation(nonce, value, gasLimit, gasPrice, byteCode))
}

func sign(tx *ethTypes.Transaction) *ethTypes.Transaction {
	privateKey, err := ethCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	return signedTx
}

func createSignedMintTx() *ethTypes.Transaction {
	generator := testutils.RandInts()
	contractAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	toAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(testutils.RandInts().Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)
	data := types.CreateMintCallData(toAddr, big.NewInt(testutils.RandIntBetween(1, math.MaxInt64)))
	return sign(ethTypes.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data))
}

func createBasicRPCMock(tx *ethTypes.Transaction, confCount int64) *ethMock.RPCClientMock {
	blockNum := testutils.RandIntBetween(confCount, 100000000)

	rpc := ethMock.RPCClientMock{
		NetworkIDFunc: func(ctx context.Context) (*big.Int, error) {
			return network.Params().ChainID, nil
		},
		TransactionByHashFunc: func(ctx context.Context, hash common.Hash) (_ *ethTypes.Transaction, isPending bool, err error) {
			if bytes.Equal(tx.Hash().Bytes(), hash.Bytes()) {
				return tx, false, nil
			}
			return nil, false, fmt.Errorf("transaction not found")
		},
		TransactionReceiptFunc: func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
			if bytes.Equal(tx.Hash().Bytes(), hash.Bytes()) {
				return &ethTypes.Receipt{BlockNumber: big.NewInt(blockNum - confCount)}, nil
			}
			return nil, fmt.Errorf("transaction not found")
		},
		BlockNumberFunc: func(ctx context.Context) (uint64, error) {
			return big.NewInt(blockNum).Uint64(), nil
		},
	}

	return &rpc
}

func createVoterMock() *ethMock.VoterMock {
	return &ethMock.VoterMock{
		InitPollFunc:   func(sdk.Context, vote.PollMeta) error { return nil },
		RecordVoteFunc: func(sdk.Context, vote.MsgVote) error { return nil },
	}
}

func assertVotedOnPoll(t *testing.T, voter *ethMock.VoterMock, hash common.Hash, verified bool) {

	assert.Equal(t, 1, len(voter.InitPollCalls()))
	assert.Equal(t, types.ModuleName, voter.InitPollCalls()[0].Poll.Module)
	assert.Equal(t, types.MsgVerifyTx{}.Type(), voter.InitPollCalls()[0].Poll.Type)
	assert.Equal(t, hash.String(), voter.InitPollCalls()[0].Poll.ID)

	initPoll := voter.InitPollCalls()[0].Poll

	assert.Equal(t, 1, len(voter.RecordVoteCalls()))
	assert.Equal(t, initPoll, voter.RecordVoteCalls()[0].Vote.Poll())
	assert.Equal(t, verified, voter.RecordVoteCalls()[0].Vote.Data())

}

func newKeeper(ctx sdk.Context, confHeight int64) keeper.Keeper {
	cdc := testutils.Codec()
	subspace := params.NewSubspace(cdc, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"), "sub")
	k := keeper.NewEthKeeper(cdc, sdk.NewKVStoreKey("testKey"), subspace)
	k.SetParams(ctx, types.Params{Network: network, ConfirmationHeight: uint64(confHeight)})
	return k
}
