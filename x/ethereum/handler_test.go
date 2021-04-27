package ethereum

import (
	"fmt"
	"math/big"
	mathRand "math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/exported"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	ethMock "github.com/axelarnetwork/axelar-core/x/ethereum/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

const (
	network = types.Network(types.Rinkeby)
)

var (
	bytecodes = common.FromHex(MymintableBin)
	tokenBC   = rand.Bytes(64)
	burnerBC  = rand.Bytes(64)
	gateway   = "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7"
)

func TestLink_NoGateway(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encCfg := testutils.MakeEncodingConfig()

	subspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"), "sub")
	k := keeper.NewEthKeeper(encCfg.Amino, sdk.NewKVStoreKey("testKey"), subspace)
	k.SetParams(ctx, types.Params{Network: network, ConfirmationHeight: uint64(minConfHeight), Gateway: bytecodes, Token: tokenBC, Burnable: burnerBC, RevoteLockingPeriod: 50})

	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	symbol := rand.Str(3)

	n := &ethMock.NexusMock{}
	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
			return rand.StrBetween(5, 20), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	handler := NewHandler(k, &ethMock.VoterMock{}, signer, n, &ethMock.SnapshotterMock{})
	_, err := handler(ctx, &types.MsgLink{Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, Symbol: symbol, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 0, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 0, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_NoRecipientChain(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)

	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	symbol := rand.Str(3)

	chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
	n := &ethMock.NexusMock{
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}

	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
			return rand.StrBetween(5, 20), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	handler := NewHandler(k, &ethMock.VoterMock{}, signer, n, &ethMock.SnapshotterMock{})
	_, err := handler(ctx, &types.MsgLink{Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, Symbol: symbol, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 0, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_NoRegisteredAsset(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)

	symbol := rand.Str(3)

	chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
	n := &ethMock.NexusMock{
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
		IsAssetRegisteredFunc: func(_ sdk.Context, chainName, denom string) bool { return false },
	}

	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
			return rand.StrBetween(5, 20), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	handler := NewHandler(k, &ethMock.VoterMock{}, signer, n, &ethMock.SnapshotterMock{})
	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	_, err := handler(ctx, &types.MsgLink{Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, Symbol: symbol, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_Success(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, minConfHeight)
	msg := createMsgSignDeploy()

	k.SetTokenInfo(ctx, msg)

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
		IsAssetRegisteredFunc: func(_ sdk.Context, chainName, denom string) bool { return true },
	}
	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
			return rand.StrBetween(5, 20), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	handler := NewHandler(k, &ethMock.VoterMock{}, signer, n, &ethMock.SnapshotterMock{})
	_, err = handler(ctx, &types.MsgLink{Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Symbol: msg.Symbol})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)

	assert.Equal(t, types.BurnerInfo{TokenAddr: tokenAddr.Hex(), Symbol: msg.Symbol, Salt: salt}, *k.GetBurnerInfo(ctx, burnAddr))
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
	newValue := big.NewInt(rand.I64Between(1, 10000))
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
	newData := rand.Bytes(int(rand.I64Between(1, 10000)))
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
	newValue := big.NewInt(rand.I64Between(1, 10000))
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
	newData := rand.Bytes(int(rand.I64Between(1, 10000)))
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
	newTo := common.BytesToAddress(rand.Bytes(common.AddressLength))
	tx2 := sign(ethTypes.NewTransaction(tx1.Nonce(), newTo, tx1.Value(), tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(network.Params().ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestHandleMsgConfirmTokenDeploy(t *testing.T) {
	var (
		ctx sdk.Context
		k   *ethMock.EthKeeperMock
		v   *ethMock.VoterMock
		n   *ethMock.NexusMock
		s   *ethMock.SignerMock
		msg *types.MsgConfirmToken
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		encCfg := testutils.MakeEncodingConfig()

		k = &ethMock.EthKeeperMock{
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), true
			},
			GetTokenAddressFunc: func(sdk.Context, string, common.Address) (common.Address, error) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), nil
			},
			GetRevoteLockingPeriodFunc:        func(ctx sdk.Context) int64 { return rand.PosI64() },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) uint64 { return mathRand.Uint64() },
			SetPendingTokenDeployFunc:         func(sdk.Context, vote.PollMeta, types.ERC20TokenDeploy) {},
			CodecFunc:                         func() *codec.LegacyAmino { return encCfg.Amino },
		}
		v = &ethMock.VoterMock{InitPollFunc: func(sdk.Context, vote.PollMeta, int64) error { return nil }}
		n = &ethMock.NexusMock{IsAssetRegisteredFunc: func(sdk.Context, string, string) bool { return false }}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
				return rand.StrBetween(5, 20), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
		}

		msg = &types.MsgConfirmToken{
			Sender: rand.Bytes(20),
			TxID:   common.BytesToHash(rand.Bytes(common.HashLength)).Hex(),
			Symbol: rand.StrBetween(5, 10),
		}
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		result, err := HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(result.Events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeTokenConfirmation }), 1)
		assert.Equal(t, v.InitPollCalls()[0].Poll, k.SetPendingTokenDeployCalls()[0].Poll)
	}).Repeat(repeats))

	t.Run("no gateway", testutils.Func(func(t *testing.T) {
		setup()
		k.GetGatewayAddressFunc = func(sdk.Context) (common.Address, bool) { return common.Address{}, false }

		_, err := HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("token unknown", testutils.Func(func(t *testing.T) {
		setup()
		k.GetTokenAddressFunc = func(sdk.Context, string, common.Address) (common.Address, error) {
			return common.Address{}, fmt.Errorf("failed")
		}

		_, err := HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("already registered", testutils.Func(func(t *testing.T) {
		setup()
		n.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return true }

		_, err := HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitPollFunc = func(sdk.Context, vote.PollMeta, int64) error { return fmt.Errorf("poll setup failed") }

		_, err := HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no key", testutils.Func(func(t *testing.T) {
		setup()
		s.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (string, bool) { return "", false }

		_, err := HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no snapshot counter", testutils.Func(func(t *testing.T) {
		setup()
		s.GetSnapshotCounterForKeyIDFunc = func(sdk.Context, string) (int64, bool) { return 0, false }

		_, err := HandleMsgConfirmTokenDeploy(ctx, k, v, s, n, msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgConfirmDeposit(t *testing.T) {
	var (
		ctx sdk.Context
		k   *ethMock.EthKeeperMock
		v   *ethMock.VoterMock
		s   *ethMock.SignerMock
		msg *types.MsgConfirmDeposit
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		encCfg := testutils.MakeEncodingConfig()
		k = &ethMock.EthKeeperMock{
			GetDepositFunc: func(sdk.Context, string, string) (types.ERC20Deposit, types.DepositState, bool) {
				return types.ERC20Deposit{}, 0, false
			},
			GetBurnerInfoFunc: func(sdk.Context, common.Address) *types.BurnerInfo {
				return &types.BurnerInfo{
					TokenAddr: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
					Symbol:    rand.StrBetween(5, 10),
					Salt:      common.BytesToHash(rand.Bytes(common.HashLength)),
				}
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) int64 { return rand.PosI64() },
			SetPendingDepositFunc:             func(sdk.Context, vote.PollMeta, *types.ERC20Deposit) {},
			GetRequiredConfirmationHeightFunc: func(sdk.Context) uint64 { return mathRand.Uint64() },
			CodecFunc:                         func() *codec.LegacyAmino { return encCfg.Amino },
		}
		v = &ethMock.VoterMock{InitPollFunc: func(sdk.Context, vote.PollMeta, int64) error { return nil }}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
				return rand.StrBetween(5, 20), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
		}

		msg = &types.MsgConfirmDeposit{
			Sender:     rand.Bytes(20),
			TxID:       common.BytesToHash(rand.Bytes(common.HashLength)).Hex(),
			Amount:     sdk.NewUint(mathRand.Uint64()),
			BurnerAddr: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
		}
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		result, err := HandleMsgConfirmDeposit(ctx, k, v, s, msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(result.Events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeDepositConfirmation }), 1)
		assert.Equal(t, v.InitPollCalls()[0].Poll, k.SetPendingDepositCalls()[0].Poll)
	}).Repeat(repeats))

	t.Run("deposit confirmed", testutils.Func(func(t *testing.T) {
		setup()
		k.GetDepositFunc = func(sdk.Context, string, string) (types.ERC20Deposit, types.DepositState, bool) {
			return types.ERC20Deposit{
				TxID:       common.BytesToHash(rand.Bytes(common.HashLength)),
				Amount:     sdk.NewUint(mathRand.Uint64()),
				Symbol:     rand.StrBetween(5, 10),
				BurnerAddr: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
			}, types.CONFIRMED, true
		}

		_, err := HandleMsgConfirmDeposit(ctx, k, v, s, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("deposit burned", testutils.Func(func(t *testing.T) {
		setup()
		k.GetDepositFunc = func(sdk.Context, string, string) (types.ERC20Deposit, types.DepositState, bool) {
			return types.ERC20Deposit{
				TxID:       common.BytesToHash(rand.Bytes(common.HashLength)),
				Amount:     sdk.NewUint(mathRand.Uint64()),
				Symbol:     rand.StrBetween(5, 10),
				BurnerAddr: common.BytesToAddress(rand.Bytes(common.AddressLength)).Hex(),
			}, types.BURNED, true
		}

		_, err := HandleMsgConfirmDeposit(ctx, k, v, s, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("burner address unknown", testutils.Func(func(t *testing.T) {
		setup()
		k.GetBurnerInfoFunc = func(sdk.Context, common.Address) *types.BurnerInfo { return nil }

		_, err := HandleMsgConfirmDeposit(ctx, k, v, s, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitPollFunc = func(sdk.Context, vote.PollMeta, int64) error { return fmt.Errorf("failed") }

		_, err := HandleMsgConfirmDeposit(ctx, k, v, s, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no key", testutils.Func(func(t *testing.T) {
		setup()
		s.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (string, bool) { return "", false }

		_, err := HandleMsgConfirmDeposit(ctx, k, v, s, msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no snapshot counter", testutils.Func(func(t *testing.T) {
		setup()
		s.GetSnapshotCounterForKeyIDFunc = func(sdk.Context, string) (int64, bool) { return 0, false }

		_, err := HandleMsgConfirmDeposit(ctx, k, v, s, msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func createSignedDeployTx() *ethTypes.Transaction {
	generator := rand.PInt64Gen()

	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(generator.Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)
	byteCode := rand.Bytes(int(rand.I64Between(1, 10000)))

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
	generator := rand.PInt64Gen()
	contractAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))
	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(rand.PInt64Gen().Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)

	data := rand.Bytes(int(rand.I64Between(0, 1000)))
	return sign(ethTypes.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data))
}

func newKeeper(ctx sdk.Context, confHeight int64) keeper.Keeper {
	encCfg := testutils.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"), "sub")
	k := keeper.NewEthKeeper(encCfg.Amino, sdk.NewKVStoreKey("testKey"), subspace)
	k.SetParams(ctx, types.Params{Network: network, ConfirmationHeight: uint64(confHeight), Gateway: bytecodes, Token: tokenBC, Burnable: burnerBC, RevoteLockingPeriod: 50})
	k.SetGatewayAddress(ctx, common.HexToAddress(gateway))

	return k
}

func createMsgSignDeploy() *types.MsgSignDeployToken {
	account := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
	symbol := rand.Str(3)
	name := rand.Str(10)
	decimals := rand.Bytes(1)[0]
	capacity := sdk.NewIntFromUint64(uint64(rand.PosI64()))

	return &types.MsgSignDeployToken{Sender: account, TokenName: name, Symbol: symbol, Decimals: decimals, Capacity: capacity}
}
