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
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	balance "github.com/axelarnetwork/axelar-core/x/balance/exported"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/exported"
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

var (
	sender   = sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 20))))
	tokenBC  = testutils.RandBytes(64)
	burnerBC = testutils.RandBytes(64)
)

func TestLink_NoSymbolSet(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)

	recipient := balance.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	symbol := testutils.RandString(3)
	gateway := "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7"

	handler := NewHandler(k, &ethMock.RPCClientMock{}, &ethMock.VoterMock{}, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.BalancerMock{})
	_, err := handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), RecipientAddr: recipient.Address, Symbol: symbol, GatewayAddr: gateway, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
}

func TestLink_Success(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)

	account, err := sdk.AccAddressFromBech32("cosmos1vjyc4qmsdtdl5a4ruymnjqpchm5gyqde63sqdh")
	if err != nil {
		panic(err)
	}

	symbol := testutils.RandString(3)
	name := testutils.RandString(10)
	decimals := testutils.RandBytes(1)[0]
	capacity := sdk.NewIntFromUint64(uint64(testutils.RandPosInt()))
	gateway := "0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA"
	k.SaveTokenInfo(ctx, types.MsgSignDeployToken{Sender: account, TokenName: name, Symbol: symbol, Decimals: decimals, Capacity: capacity})

	recipient := balance.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	burnAddr, salt, err := k.GetBurnerAddressAndSalt(ctx, symbol, recipient.Address, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}
	sender := balance.CrossChainAddress{Address: burnAddr.String(), Chain: exported.Ethereum}

	chains := map[string]balance.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
	b := &ethMock.BalancerMock{
		LinkAddressesFunc: func(ctx sdk.Context, s balance.CrossChainAddress, r balance.CrossChainAddress) {},
		GetChainFunc: func(ctx sdk.Context, chain string) (balance.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}
	handler := NewHandler(k, &ethMock.RPCClientMock{}, &ethMock.VoterMock{}, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, b)
	_, err = handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Symbol: symbol, GatewayAddr: gateway})

	assert.NoError(t, err)

	assert.Equal(t, 1, len(b.LinkAddressesCalls()))
	assert.Equal(t, sender, b.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, b.LinkAddressesCalls()[0].Recipient)

	assert.Equal(t, types.BurnerInfo{Symbol: symbol, Salt: salt}, *k.GetBurnerInfo(ctx, burnAddr))
}

func TestDeployTx_DifferentValue_DifferentHash(t *testing.T) {
	tx1 := createSignedDeployTx()
	privateKey, err := ethCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newValue := big.NewInt(testutils.RandIntBetween(1, 10000))
	tx2 := sign(ethTypes.NewContractCreation(tx1.Nonce(), newValue, tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestDeployTx_DifferentData_DifferentHash(t *testing.T) {
	tx1 := createSignedDeployTx()
	privateKey, err := ethCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newData := testutils.RandBytes(int(testutils.RandIntBetween(1, 10000)))
	tx2 := sign(ethTypes.NewContractCreation(tx1.Nonce(), tx1.Value(), tx1.Gas(), tx1.GasPrice(), newData))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestMintTx_DifferentValue_DifferentHash(t *testing.T) {
	tx1 := createSignedEthTx()
	privateKey, err := ethCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newValue := big.NewInt(testutils.RandIntBetween(1, 10000))
	tx2 := sign(ethTypes.NewTransaction(tx1.Nonce(), *tx1.To(), newValue, tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestMintTx_DifferentData_DifferentHash(t *testing.T) {
	tx1 := createSignedEthTx()
	privateKey, err := ethCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newData := testutils.RandBytes(int(testutils.RandIntBetween(1, 10000)))
	tx2 := sign(ethTypes.NewTransaction(tx1.Nonce(), *tx1.To(), tx1.Value(), tx1.Gas(), tx1.GasPrice(), newData))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestMintTx_DifferentRecipient_DifferentHash(t *testing.T) {
	tx1 := createSignedEthTx()
	privateKey, err := ethCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newTo := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	tx2 := sign(ethTypes.NewTransaction(tx1.Nonce(), newTo, tx1.Value(), tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestVerifyTx_Deploy_HashNotFound(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedDeployTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
		return nil, fmt.Errorf("wrong hash")
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, createSnapshotter(), &ethMock.BalancerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, types.ModuleCdc.MustMarshalJSON(signedTx)))

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
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.BalancerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, types.ModuleCdc.MustMarshalJSON(signedTx)))

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
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.BalancerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, types.ModuleCdc.MustMarshalJSON(signedTx)))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), true)
}

func TestVerifyTx_Mint_HashNotFound(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	rpc.TransactionReceiptFunc = func(ctx context.Context, txHash common.Hash) (*ethTypes.Receipt, error) {
		return nil, fmt.Errorf("wrong hash")
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.BalancerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, types.ModuleCdc.MustMarshalJSON(signedTx)))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_NotConfirmed(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(0, minConfHeight)
	signedTx := createSignedEthTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.BalancerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, types.ModuleCdc.MustMarshalJSON(signedTx)))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), false)
}

func TestVerifyTx_Mint_Success(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.BalancerMock{})

	_, err := handler(ctx, types.NewMsgVerifyTx(sender, types.ModuleCdc.MustMarshalJSON(signedTx)))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedTx(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), true)
}

func createSignedDeployTx() *ethTypes.Transaction {
	generator := testutils.RandPosInts()

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

func createSignedEthTx() *ethTypes.Transaction {
	generator := testutils.RandPosInts()
	contractAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(testutils.RandPosInts().Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)

	data := testutils.RandBytes(int(testutils.RandIntBetween(0, 1000)))
	return sign(ethTypes.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data))
}

func createBasicRPCMock(tx *ethTypes.Transaction, confCount int64) *ethMock.RPCClientMock {
	blockNum := testutils.RandIntBetween(confCount, 100000000)

	rpc := ethMock.RPCClientMock{
		ChainIDFunc: func(ctx context.Context) (*big.Int, error) {
			return network.Params().ChainID, nil
		},
		TransactionReceiptFunc: func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
			if bytes.Equal(tx.Hash().Bytes(), hash.Bytes()) {
				return &ethTypes.Receipt{TxHash: tx.Hash(), BlockNumber: big.NewInt(blockNum - confCount)}, nil
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
	k.SetParams(ctx, types.Params{Network: network, ConfirmationHeight: uint64(confHeight), Token: tokenBC, Burnable: burnerBC})
	return k
}

func createSnapshotter() types.Snapshotter {
	return &ethMock.SnapshotterMock{}
}
