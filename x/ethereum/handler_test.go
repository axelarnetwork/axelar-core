package ethereum

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	ethMock "github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
	sender      = sdk.AccAddress(testutils.RandString(int(testutils.RandIntBetween(5, 20))))
	bytecodes   = common.FromHex(MymintableBin)
	tokenBC     = testutils.RandBytes(64)
	burnerBC    = testutils.RandBytes(64)
	transferSig = testutils.RandBytes(64)
	gateway     = "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7"
)

func TestLink_NoSymbolSet(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)

	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	symbol := testutils.RandString(3)

	handler := NewHandler(k, &ethMock.RPCClientMock{}, &ethMock.VoterMock{}, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})
	_, err := handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), RecipientAddr: recipient.Address, Symbol: symbol, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
}

func TestLink_Success(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	msg := createMsgSignDeploy()

	k.SaveTokenInfo(ctx, msg)

	recipient := nexus.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}

	burnAddr, salt, err := k.GetBurnerAddressAndSalt(ctx, tokenAddr, recipient.Address, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}
	sender := nexus.CrossChainAddress{Address: burnAddr.String(), Chain: exported.Ethereum}

	chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
	n := &ethMock.NexusMock{
		LinkAddressesFunc: func(ctx sdk.Context, s nexus.CrossChainAddress, r nexus.CrossChainAddress) {},
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}
	handler := NewHandler(k, &ethMock.RPCClientMock{}, &ethMock.VoterMock{}, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, n)
	_, err = handler(ctx, types.MsgLink{Sender: sdk.AccAddress("sender"), RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Symbol: msg.Symbol})

	assert.NoError(t, err)

	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)

	assert.Equal(t, types.BurnerInfo{TokenAddr: tokenAddr.String(), Symbol: msg.Symbol, Salt: salt}, *k.GetBurnerInfo(ctx, burnAddr))
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

func TestVerifyToken_NoTokenInfo(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()
	symbol := testutils.RandString(4)

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := createBasicRPCMock(signedTx, confCount, nil)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, createSnapshotter(), &ethMock.NexusMock{})

	_, err := handler(ctx, types.NewMsgVerifyErc20TokenDeploy(sender, signedTx.Hash(), symbol))

	assert.Error(t, err)
	assert.False(t, k.HasUnverifiedToken(ctx, signedTx.Hash().String()))
	assert.Equal(t, 0, len(voter.InitPollCalls()))
	assert.Equal(t, 0, len(voter.RecordVoteCalls()))
}

func TestVerifyToken_NoReceipt(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()
	msg := createMsgSignDeploy()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SaveTokenInfo(ctx, msg)
	rpc := createBasicRPCMock(signedTx, confCount, nil)
	rpc.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
		return nil, fmt.Errorf("no transaction for hash")
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, createSnapshotter(), &ethMock.NexusMock{})

	_, err := handler(ctx, types.NewMsgVerifyErc20TokenDeploy(sender, signedTx.Hash(), msg.Symbol))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedToken(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), types.MsgVerifyErc20TokenDeploy{}.Type(), false)
}

func TestVerifyToken_NoBlockNumber(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()
	msg := createMsgSignDeploy()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SaveTokenInfo(ctx, msg)
	rpc := createBasicRPCMock(signedTx, confCount, nil)
	rpc.BlockNumberFunc = func(ctx context.Context) (uint64, error) {
		return 0, fmt.Errorf("no block number")
	}
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, createSnapshotter(), &ethMock.NexusMock{})

	_, err := handler(ctx, types.NewMsgVerifyErc20TokenDeploy(sender, signedTx.Hash(), msg.Symbol))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedToken(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), types.MsgVerifyErc20TokenDeploy{}.Type(), false)
}

func TestVerifyToken_NotConfirmed(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(0, minConfHeight)
	signedTx := createSignedEthTx()
	msg := createMsgSignDeploy()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SaveTokenInfo(ctx, msg)
	rpc := createBasicRPCMock(signedTx, confCount, nil)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	_, err := handler(ctx, types.NewMsgVerifyErc20TokenDeploy(sender, signedTx.Hash(), msg.Symbol))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedToken(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), types.MsgVerifyErc20TokenDeploy{}.Type(), false)
}

func TestVerifyToken_NoEvent(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()
	msg := createMsgSignDeploy()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SaveTokenInfo(ctx, msg)
	logs := createLogs("", common.Address{}, common.Address{}, common.Hash{}, false)
	rpc := createBasicRPCMock(signedTx, confCount, logs)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	_, err := handler(ctx, types.NewMsgVerifyErc20TokenDeploy(sender, signedTx.Hash(), msg.Symbol))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedToken(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), types.MsgVerifyErc20TokenDeploy{}.Type(), false)
}
func TestVerifyToken_DifferentEvent(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()
	msg := createMsgSignDeploy()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SaveTokenInfo(ctx, msg)
	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}
	logs := createLogs(testutils.RandString(4), common.HexToAddress(gateway), tokenAddr, k.GetERC20TokenDeploySignature(ctx), true)
	rpc := createBasicRPCMock(signedTx, confCount, logs)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	_, err = handler(ctx, types.NewMsgVerifyErc20TokenDeploy(sender, signedTx.Hash(), msg.Symbol))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedToken(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), types.MsgVerifyErc20TokenDeploy{}.Type(), false)
}

func TestVerifyToken_Success(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	confCount := testutils.RandIntBetween(minConfHeight, 10*minConfHeight)
	signedTx := createSignedEthTx()
	msg := createMsgSignDeploy()

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SaveTokenInfo(ctx, msg)
	tokenAddr, err := k.GetTokenAddress(ctx, msg.Symbol, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}
	logs := createLogs(msg.Symbol, common.HexToAddress(gateway), tokenAddr, k.GetERC20TokenDeploySignature(ctx), true)
	rpc := createBasicRPCMock(signedTx, confCount, logs)
	voter := createVoterMock()
	handler := NewHandler(k, rpc, voter, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	_, err = handler(ctx, types.NewMsgVerifyErc20TokenDeploy(sender, signedTx.Hash(), msg.Symbol))

	assert.NoError(t, err)
	assert.True(t, k.HasUnverifiedToken(ctx, signedTx.Hash().String()))
	assertVotedOnPoll(t, voter, signedTx.Hash(), types.MsgVerifyErc20TokenDeploy{}.Type(), true)
}

func TestHandleMsgVerifyErc20Deposit_UnknownBurnerAddr(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	txID := common.BytesToHash(testutils.RandBytes(common.HashLength))
	amount := sdk.NewUint(uint64(testutils.RandIntBetween(1, 10000)))
	unknownBurnerAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	rpc := ethMock.RPCClientMock{}
	v := createVoterMock()
	handler := NewHandler(k, &rpc, v, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	msg := types.NewMsgVerifyErc20Deposit(sender, txID, amount, unknownBurnerAddr)
	result, err := handler(ctx, msg)

	assert.Nil(t, result)
	assert.Error(t, err)

	assert.False(t, k.HasUnverifiedErc20Deposit(ctx, txID.String()))
	assert.Equal(t, 0, len(v.InitPollCalls()))
	assert.Equal(t, 0, len(v.RecordVoteCalls()))
}

func TestHandleMsgVerifyErc20Deposit_FailedGettingTransactionReceipt(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	txID := common.BytesToHash(testutils.RandBytes(common.HashLength))
	amount := sdk.NewUint(uint64(testutils.RandIntBetween(1, 10000)))
	tokenAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	symbol := testutils.RandString(3)
	salt := common.BytesToHash(testutils.RandBytes(common.HashLength))
	burnerAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	burnerInfo := types.BurnerInfo{
		TokenAddr: tokenAddr.String(),
		Symbol:    symbol,
		Salt:      salt,
	}

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)
	rpc := ethMock.RPCClientMock{}
	rpc.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
		return nil, fmt.Errorf("sorry")
	}
	v := createVoterMock()
	handler := NewHandler(k, &rpc, v, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	msg := types.NewMsgVerifyErc20Deposit(sender, txID, amount, burnerAddr)
	result, err := handler(ctx, msg)

	assert.NotNil(t, result)
	assert.NoError(t, err)

	assert.True(t, k.HasUnverifiedErc20Deposit(ctx, txID.String()))
	assertVotedOnPoll(t, v, txID, types.MsgVerifyErc20Deposit{}.Type(), false)
}

func TestHandleMsgVerifyErc20Deposit_FailedGettingBlockNumber(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	txID := common.BytesToHash(testutils.RandBytes(common.HashLength))
	amount := sdk.NewUint(uint64(testutils.RandIntBetween(1, 10000)))
	tokenAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	symbol := testutils.RandString(3)
	salt := common.BytesToHash(testutils.RandBytes(common.HashLength))
	burnerAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	burnerInfo := types.BurnerInfo{
		TokenAddr: tokenAddr.String(),
		Symbol:    symbol,
		Salt:      salt,
	}

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)
	rpc := ethMock.RPCClientMock{}
	rpc.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
		return &ethTypes.Receipt{}, nil
	}
	rpc.BlockNumberFunc = func(ctx context.Context) (uint64, error) {
		return 0, fmt.Errorf("sorry")
	}
	v := createVoterMock()
	handler := NewHandler(k, &rpc, v, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	msg := types.NewMsgVerifyErc20Deposit(sender, txID, amount, burnerAddr)
	result, err := handler(ctx, msg)

	assert.NotNil(t, result)
	assert.NoError(t, err)

	assert.True(t, k.HasUnverifiedErc20Deposit(ctx, txID.String()))
	assertVotedOnPoll(t, v, txID, types.MsgVerifyErc20Deposit{}.Type(), false)
}

func TestHandleMsgVerifyErc20Deposit_NotConfirmed(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	txID := common.BytesToHash(testutils.RandBytes(common.HashLength))
	amount := sdk.NewUint(uint64(testutils.RandIntBetween(1, 10000)))
	tokenAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	symbol := testutils.RandString(3)
	salt := common.BytesToHash(testutils.RandBytes(common.HashLength))
	burnerAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	burnerInfo := types.BurnerInfo{
		TokenAddr: tokenAddr.String(),
		Symbol:    symbol,
		Salt:      salt,
	}
	blockNumber := int64(100)

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)
	rpc := ethMock.RPCClientMock{}
	rpc.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
		return &ethTypes.Receipt{BlockNumber: big.NewInt(blockNumber)}, nil
	}
	rpc.BlockNumberFunc = func(ctx context.Context) (uint64, error) {
		return uint64(blockNumber), nil
	}
	v := createVoterMock()
	handler := NewHandler(k, &rpc, v, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	msg := types.NewMsgVerifyErc20Deposit(sender, txID, amount, burnerAddr)
	result, err := handler(ctx, msg)

	assert.NotNil(t, result)
	assert.NoError(t, err)

	assert.True(t, k.HasUnverifiedErc20Deposit(ctx, txID.String()))
	assertVotedOnPoll(t, v, txID, types.MsgVerifyErc20Deposit{}.Type(), false)
}

func TestHandleMsgVerifyErc20Deposit_AmountMismatch(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	txID := common.BytesToHash(testutils.RandBytes(common.HashLength))
	amount := sdk.NewUint(10)
	tokenAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	symbol := testutils.RandString(3)
	salt := common.BytesToHash(testutils.RandBytes(common.HashLength))
	burnerAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	burnerInfo := types.BurnerInfo{
		TokenAddr: tokenAddr.String(),
		Symbol:    symbol,
		Salt:      salt,
	}
	blockNumber := int64(100)
	erc20TransferEventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)
	rpc := ethMock.RPCClientMock{}
	rpc.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
		return &ethTypes.Receipt{BlockNumber: big.NewInt(blockNumber), Logs: []*ethTypes.Log{
			/* ERC20 transfer to burner address of a random token */
			{
				Address: common.BytesToAddress(testutils.RandBytes(common.AddressLength)),
				Topics: []common.Hash{
					erc20TransferEventSig,
					common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(testutils.RandBytes(common.AddressLength)).Bytes(), common.HashLength)),
					common.BytesToHash(common.LeftPadBytes(burnerAddr.Bytes(), common.HashLength)),
				},
				Data: common.LeftPadBytes(big.NewInt(2).Bytes(), common.HashLength),
			},
			/* not a ERC20 transfer */
			{
				Address: tokenAddr,
				Topics: []common.Hash{
					common.BytesToHash(testutils.RandBytes(common.HashLength)),
					common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(testutils.RandBytes(common.AddressLength)).Bytes(), common.HashLength)),
					common.BytesToHash(common.LeftPadBytes(burnerAddr.Bytes(), common.HashLength)),
				},
				Data: common.LeftPadBytes(big.NewInt(2).Bytes(), common.HashLength),
			},
			/* an invalid ERC20 transfer */
			{
				Address: tokenAddr,
				Topics: []common.Hash{
					erc20TransferEventSig,
					common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(testutils.RandBytes(common.AddressLength)).Bytes(), common.HashLength)),
				},
				Data: common.LeftPadBytes(big.NewInt(2).Bytes(), common.HashLength),
			},
			/* an ERC20 transfer of our concern */
			{
				Address: tokenAddr,
				Topics: []common.Hash{
					erc20TransferEventSig,
					common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(testutils.RandBytes(common.AddressLength)).Bytes(), common.HashLength)),
					common.BytesToHash(common.LeftPadBytes(burnerAddr.Bytes(), common.HashLength)),
				},
				Data: common.LeftPadBytes(big.NewInt(4).Bytes(), common.HashLength),
			},
		}}, nil
	}
	rpc.BlockNumberFunc = func(ctx context.Context) (uint64, error) {
		return uint64(blockNumber + minConfHeight*2), nil
	}
	v := createVoterMock()
	handler := NewHandler(k, &rpc, v, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	msg := types.NewMsgVerifyErc20Deposit(sender, txID, amount, burnerAddr)
	result, err := handler(ctx, msg)

	assert.NotNil(t, result)
	assert.NoError(t, err)

	assert.True(t, k.HasUnverifiedErc20Deposit(ctx, txID.String()))
	assertVotedOnPoll(t, v, txID, types.MsgVerifyErc20Deposit{}.Type(), false)
}

func TestHandleMsgVerifyErc20Deposit_Success(t *testing.T) {
	minConfHeight := testutils.RandIntBetween(1, 10)
	txID := common.BytesToHash(testutils.RandBytes(common.HashLength))
	amount := sdk.NewUint(10)
	tokenAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	symbol := testutils.RandString(3)
	salt := common.BytesToHash(testutils.RandBytes(common.HashLength))
	burnerAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
	burnerInfo := types.BurnerInfo{
		TokenAddr: tokenAddr.String(),
		Symbol:    symbol,
		Salt:      salt,
	}
	blockNumber := int64(100)
	erc20TransferEventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	k.SetBurnerInfo(ctx, burnerAddr, &burnerInfo)
	rpc := ethMock.RPCClientMock{}
	rpc.TransactionReceiptFunc = func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
		return &ethTypes.Receipt{BlockNumber: big.NewInt(blockNumber), Logs: []*ethTypes.Log{
			/* an ERC20 transfer of our concern */
			{
				Address: tokenAddr,
				Topics: []common.Hash{
					erc20TransferEventSig,
					common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(testutils.RandBytes(common.AddressLength)).Bytes(), common.HashLength)),
					common.BytesToHash(common.LeftPadBytes(burnerAddr.Bytes(), common.HashLength)),
				},
				Data: common.LeftPadBytes(big.NewInt(3).Bytes(), common.HashLength),
			},
			/* another ERC20 transfer of our concern */
			{
				Address: tokenAddr,
				Topics: []common.Hash{
					erc20TransferEventSig,
					common.BytesToHash(common.LeftPadBytes(common.BytesToAddress(testutils.RandBytes(common.AddressLength)).Bytes(), common.HashLength)),
					common.BytesToHash(common.LeftPadBytes(burnerAddr.Bytes(), common.HashLength)),
				},
				Data: common.LeftPadBytes(big.NewInt(7).Bytes(), common.HashLength),
			},
		}}, nil
	}
	rpc.BlockNumberFunc = func(ctx context.Context) (uint64, error) {
		return uint64(blockNumber + minConfHeight*2), nil
	}
	v := createVoterMock()
	handler := NewHandler(k, &rpc, v, &ethMock.SignerMock{}, &ethMock.SnapshotterMock{}, &ethMock.NexusMock{})

	msg := types.NewMsgVerifyErc20Deposit(sender, txID, amount, burnerAddr)
	result, err := handler(ctx, msg)

	assert.NotNil(t, result)
	assert.NoError(t, err)

	assert.True(t, k.HasUnverifiedErc20Deposit(ctx, txID.String()))
	assertVotedOnPoll(t, v, txID, types.MsgVerifyErc20Deposit{}.Type(), true)
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

func createBasicRPCMock(tx *ethTypes.Transaction, confCount int64, logs []*ethTypes.Log) *ethMock.RPCClientMock {
	blockNum := testutils.RandIntBetween(confCount, 100000000)

	rpc := ethMock.RPCClientMock{
		ChainIDFunc: func(ctx context.Context) (*big.Int, error) {
			return network.Params().ChainID, nil
		},
		TransactionReceiptFunc: func(ctx context.Context, hash common.Hash) (*ethTypes.Receipt, error) {
			if bytes.Equal(tx.Hash().Bytes(), hash.Bytes()) {
				return &ethTypes.Receipt{TxHash: tx.Hash(), BlockNumber: big.NewInt(blockNum - confCount), Logs: logs}, nil
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
		RecordVoteFunc: func(vote.MsgVote) {},
	}
}

func assertVotedOnPoll(t *testing.T, voter *ethMock.VoterMock, hash common.Hash, pollType string, verified bool) {
	assert.Equal(t, 1, len(voter.InitPollCalls()))
	assert.Equal(t, types.ModuleName, voter.InitPollCalls()[0].Poll.Module)
	assert.Equal(t, pollType, voter.InitPollCalls()[0].Poll.Type)
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
	k.SetParams(ctx, types.Params{Network: network, ConfirmationHeight: uint64(confHeight), Gateway: bytecodes, Token: tokenBC, Burnable: burnerBC, TokenDeploySig: transferSig})
	_ = k.SetGatewayAddress(ctx, common.HexToAddress(gateway))

	return k
}

func createSnapshotter() types.Snapshotter {
	return &ethMock.SnapshotterMock{}
}

func createMsgSignDeploy() types.MsgSignDeployToken {
	account := sdk.AccAddress(testutils.RandBytes(sdk.AddrLen))
	symbol := testutils.RandString(3)
	name := testutils.RandString(10)
	decimals := testutils.RandBytes(1)[0]
	capacity := sdk.NewIntFromUint64(uint64(testutils.RandPosInt()))

	return types.MsgSignDeployToken{Sender: account, TokenName: name, Symbol: symbol, Decimals: decimals, Capacity: capacity}
}

func createLogs(denom string, gateway, addr common.Address, deploySig common.Hash, contains bool) []*ethTypes.Log {
	numLogs := testutils.RandIntBetween(1, 100)
	pos := testutils.RandIntBetween(0, numLogs)
	var logs []*ethTypes.Log

	for i := int64(0); i < numLogs; i++ {
		stringType, err := abi.NewType("string", "string", nil)
		if err != nil {
			panic(err)
		}
		addressType, err := abi.NewType("address", "address", nil)
		if err != nil {
			panic(err)
		}
		args := abi.Arguments{{Type: stringType}, {Type: addressType}}

		if contains && i == pos {
			data, err := args.Pack(denom, addr)
			if err != nil {
				panic(err)
			}
			logs = append(logs, &ethTypes.Log{Address: gateway, Data: data, Topics: []common.Hash{deploySig}})
			continue
		}

		randDenom := testutils.RandString(4)
		randGateway := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
		randAddr := common.BytesToAddress(testutils.RandBytes(common.AddressLength))
		randData, err := args.Pack(randDenom, randAddr)
		randTopic := common.BytesToHash(testutils.RandBytes(common.HashLength))
		if err != nil {
			panic(err)
		}
		logs = append(logs, &ethTypes.Log{Address: randGateway, Data: randData, Topics: []common.Hash{randTopic}})
	}

	return logs
}
