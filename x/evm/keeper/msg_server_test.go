package keeper_test

import (
	"fmt"
	"math/big"
	mathRand "math/rand"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
	ethParams "github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmMock "github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	evmChain    = exported.Ethereum.Name
	network     = types.Rinkeby
	networkConf = ethParams.RinkebyChainConfig
	bytecodes   = common.FromHex(MymintableBin)
	tokenBC     = rand.Bytes(64)
	burnerBC    = rand.Bytes(64)
	gateway     = "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7"
)

func TestLink_UnknownChain(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encCfg := app.MakeEncodingConfig()

	paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("testKey"), paramsK)
	k.SetParams(ctx, types.Params{
		Chain:               exported.Ethereum.Name,
		Network:             network,
		ConfirmationHeight:  uint64(minConfHeight),
		Gateway:             bytecodes,
		Token:               tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
	})

	recipient := nexus.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	symbol := rand.Str(3)

	n := &evmMock.NexusMock{
		GetChainFunc: func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false },
	}
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, &mock.SignerMock{}, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.Bytes(sdk.AddrLen), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Symbol: symbol})

	assert.Error(t, err)
	assert.Equal(t, 0, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 1, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_NoGateway(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encCfg := app.MakeEncodingConfig()

	paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("testKey"), paramsK)
	k.SetParams(ctx, types.Params{
		Chain:               exported.Ethereum.Name,
		Network:             network,
		ConfirmationHeight:  uint64(minConfHeight),
		Gateway:             bytecodes,
		Token:               tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
	})

	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	symbol := rand.Str(3)

	chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
	n := &evmMock.NexusMock{
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
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Chain: evmChain, Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, Symbol: symbol, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 0, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 1, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_NoRecipientChain(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, "Ethereum", minConfHeight)

	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	symbol := rand.Str(3)

	chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
	n := &evmMock.NexusMock{
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
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Chain: evmChain, Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, Symbol: symbol, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 0, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_NoRegisteredAsset(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, "Ethereum", minConfHeight)

	symbol := rand.Str(3)

	chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
	n := &evmMock.NexusMock{
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
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.Bytes(sdk.AddrLen), Chain: evmChain, RecipientAddr: recipient.Address, Symbol: symbol, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_Success(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	chain := "Ethereum"
	k := newKeeper(ctx, chain, minConfHeight)
	msg := createMsgSignDeploy()

	k.ForChain(ctx, chain).SetTokenInfo(ctx, msg)

	recipient := nexus.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	tokenAddr, err := k.ForChain(ctx, chain).GetTokenAddress(ctx, msg.Symbol, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}

	burnAddr, salt, err := k.ForChain(ctx, chain).GetBurnerAddressAndSalt(ctx, tokenAddr, recipient.Address, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}
	sender := nexus.CrossChainAddress{Address: burnAddr.String(), Chain: exported.Ethereum}

	chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
	n := &evmMock.NexusMock{
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
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err = server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.Bytes(sdk.AddrLen), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Symbol: msg.Symbol})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)

	assert.Equal(t, types.BurnerInfo{TokenAddress: types.Address(tokenAddr), Symbol: msg.Symbol, Salt: types.Hash(salt)}, *k.ForChain(ctx, chain).GetBurnerInfo(ctx, burnAddr))
}

func TestDeployTx_DifferentValue_DifferentHash(t *testing.T) {
	tx1 := createSignedDeployTx()
	privateKey, err := ethCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newValue := big.NewInt(rand.I64Between(1, 10000))
	tx2 := sign(ethTypes.NewContractCreation(tx1.Nonce(), newValue, tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
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
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newData := rand.Bytes(int(rand.I64Between(1, 10000)))
	tx2 := sign(ethTypes.NewContractCreation(tx1.Nonce(), tx1.Value(), tx1.Gas(), tx1.GasPrice(), newData))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
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
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newValue := big.NewInt(rand.I64Between(1, 10000))
	tx2 := sign(ethTypes.NewTransaction(tx1.Nonce(), *tx1.To(), newValue, tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
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
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newData := rand.Bytes(int(rand.I64Between(1, 10000)))
	tx2 := sign(ethTypes.NewTransaction(tx1.Nonce(), *tx1.To(), tx1.Value(), tx1.Gas(), tx1.GasPrice(), newData))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
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
	tx1, err = ethTypes.SignTx(tx1, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newTo := common.BytesToAddress(rand.Bytes(common.AddressLength))
	tx2 := sign(ethTypes.NewTransaction(tx1.Nonce(), newTo, tx1.Value(), tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = ethTypes.SignTx(tx2, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestHandleMsgConfirmChain(t *testing.T) {
	var (
		ctx    sdk.Context
		basek  *evmMock.BaseKeeperMock
		chaink *evmMock.ChainKeeperMock
		v      *evmMock.VoterMock
		n      *evmMock.NexusMock
		s      *evmMock.SnapshotterMock
		msg    *types.ConfirmChainRequest
		server types.MsgServiceServer
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		msg = &types.ConfirmChainRequest{
			Sender: rand.Bytes(20),
			Name:   rand.StrBetween(5, 20),
		}

		chaink = &evmMock.ChainKeeperMock{
			GetRevoteLockingPeriodFunc: func(ctx sdk.Context) (int64, bool) {
				return rand.I64Between(50, 100), true
			},
		}

		basek = &evmMock.BaseKeeperMock{

			ForChainFunc: func(ctx sdk.Context, chain string) types.ChainKeeper {
				if strings.ToLower(chain) == strings.ToLower(msg.Name) {
					return chaink
				}
				return nil
			},

			SetPendingChainFunc: func(sdk.Context, nexus.Chain) {},
			GetPendingChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				if strings.ToLower(chain) == strings.ToLower(msg.Name) {
					return nexus.Chain{Name: msg.Name, NativeAsset: rand.StrBetween(3, 5), SupportsForeignAssets: true}, true
				}
				return nexus.Chain{}, false
			},
		}
		v = &evmMock.VoterMock{InitPollFunc: func(sdk.Context, vote.PollMeta, int64, int64) error { return nil }}
		chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
		n = &evmMock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
			IsAssetRegisteredFunc: func(sdk.Context, string, string) bool { return false },
		}
		s = &mock.SnapshotterMock{
			GetLatestCounterFunc: func(sdk.Context) int64 {
				return rand.I64Between(50, 100)

			},
		}

		server = keeper.NewMsgServerImpl(basek, &mock.TSSMock{}, n, &mock.SignerMock{}, v, s)
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeChainConfirmation }), 1)
		assert.Equal(t, 1, len(v.InitPollCalls()))
	}).Repeat(repeats))

	t.Run("happy path with no snapshot", testutils.Func(func(t *testing.T) {
		setup()

		s = &mock.SnapshotterMock{
			GetLatestCounterFunc: func(sdk.Context) int64 {
				if len(s.TakeSnapshotCalls()) > 0 {
					return rand.I64Between(50, 100)
				}
				return -1
			},
			TakeSnapshotFunc: func(sdk.Context, int64, tss.KeyShareDistributionPolicy) (sdk.Int, sdk.Int, error) {
				return sdk.NewInt(rand.I64Between(10000, 100000)), sdk.NewInt(rand.I64Between(10000, 100000)), nil
			},
		}

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeChainConfirmation }), 1)
		assert.Equal(t, 1, len(v.InitPollCalls()))
	}).Repeat(repeats))

	t.Run("registered chain", testutils.Func(func(t *testing.T) {
		setup()
		msg.Name = evmChain

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		basek.GetPendingChainFunc = func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false }

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitPollFunc = func(sdk.Context, vote.PollMeta, int64, int64) error { return fmt.Errorf("poll setup failed") }

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgConfirmTokenDeploy(t *testing.T) {
	var (
		ctx    sdk.Context
		basek  *evmMock.BaseKeeperMock
		chaink *evmMock.ChainKeeperMock
		v      *evmMock.VoterMock
		n      *evmMock.NexusMock
		s      *evmMock.SignerMock
		msg    *types.ConfirmTokenRequest
		server types.MsgServiceServer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		basek = &evmMock.BaseKeeperMock{
			ForChainFunc: func(ctx sdk.Context, chain string) types.ChainKeeper {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return chaink
				}
				return nil
			},
		}
		chaink = &evmMock.ChainKeeperMock{
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), true
			},
			GetTokenAddressFunc: func(sdk.Context, string, common.Address) (common.Address, error) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), nil
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) (int64, bool) { return rand.PosI64(), true },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
			SetPendingTokenDeploymentFunc:     func(sdk.Context, vote.PollMeta, types.ERC20TokenDeployment) {},
		}
		v = &evmMock.VoterMock{InitPollFunc: func(sdk.Context, vote.PollMeta, int64, int64) error { return nil }}
		chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
		n = &evmMock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
			IsAssetRegisteredFunc: func(sdk.Context, string, string) bool { return false },
		}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
				return rand.StrBetween(5, 20), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
		}

		msg = &types.ConfirmTokenRequest{
			Sender: rand.Bytes(20),
			Chain:  evmChain,
			TxID:   types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Symbol: rand.StrBetween(5, 10),
		}

		server = keeper.NewMsgServerImpl(basek, &mock.TSSMock{}, n, s, v, &mock.SnapshotterMock{})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeTokenConfirmation }), 1)
		assert.Equal(t, v.InitPollCalls()[0].Poll, chaink.SetPendingTokenDeploymentCalls()[0].Poll)
	}).Repeat(repeats))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		msg.Chain = rand.StrBetween(5, 20)

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no gateway", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetGatewayAddressFunc = func(sdk.Context) (common.Address, bool) { return common.Address{}, false }

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("token unknown", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetTokenAddressFunc = func(sdk.Context, string, common.Address) (common.Address, error) {
			return common.Address{}, fmt.Errorf("failed")
		}

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("already registered", testutils.Func(func(t *testing.T) {
		setup()
		n.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return true }

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitPollFunc = func(sdk.Context, vote.PollMeta, int64, int64) error { return fmt.Errorf("poll setup failed") }

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no key", testutils.Func(func(t *testing.T) {
		setup()
		s.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (string, bool) { return "", false }

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no snapshot counter", testutils.Func(func(t *testing.T) {
		setup()
		s.GetSnapshotCounterForKeyIDFunc = func(sdk.Context, string) (int64, bool) { return 0, false }

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestAddChain(t *testing.T) {
	var (
		ctx         sdk.Context
		basek       *evmMock.BaseKeeperMock
		tssMock     *evmMock.TSSMock
		n           *evmMock.NexusMock
		msg         *types.AddChainRequest
		server      types.MsgServiceServer
		name        string
		nativeAsset string
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		chains := map[string]nexus.Chain{
			exported.Ethereum.Name: exported.Ethereum,
			btc.Bitcoin.Name:       btc.Bitcoin,
		}
		basek = &evmMock.BaseKeeperMock{
			SetParamsFunc:       func(sdk.Context, ...types.Params) {},
			SetPendingChainFunc: func(sdk.Context, nexus.Chain) {},
		}

		tssMock = &evmMock.TSSMock{
			SetKeyRequirementFunc: func(sdk.Context, tss.KeyRequirement) {},
		}

		n = &evmMock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}

		name = rand.StrBetween(5, 20)
		nativeAsset = rand.StrBetween(3, 10)
		params := types.DefaultParams()[0]
		params.Chain = name
		keyReqs := tssTypes.DefaultParams().KeyRequirements[0]
		keyReqs.ChainName = name
		msg = &types.AddChainRequest{
			Sender:         rand.Bytes(20),
			Name:           name,
			NativeAsset:    nativeAsset,
			KeyRequirement: keyReqs,
			Params:         params,
		}

		server = keeper.NewMsgServerImpl(basek, tssMock, n, &evmMock.SignerMock{}, &evmMock.VoterMock{}, &mock.SnapshotterMock{})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(basek.SetParamsCalls()))
		assert.Equal(t, 1, len(tssMock.SetKeyRequirementCalls()))
		assert.Equal(t, 1, len(basek.SetPendingChainCalls()))
		assert.Equal(t, name, basek.SetPendingChainCalls()[0].Chain.Name)
		assert.Equal(t, nativeAsset, basek.SetPendingChainCalls()[0].Chain.NativeAsset)
		assert.Equal(t, name, basek.SetParamsCalls()[0].Params[0].Chain)
		assert.Equal(t, name, tssMock.SetKeyRequirementCalls()[0].KeyRequirement.ChainName)

		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeNewChain }), 1)

	}).Repeat(repeats))

	t.Run("chain already registered", testutils.Func(func(t *testing.T) {
		setup()

		msg.Name = "Bitcoin"
		msg.NativeAsset = nativeAsset

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgConfirmDeposit(t *testing.T) {
	var (
		ctx    sdk.Context
		basek  *evmMock.BaseKeeperMock
		chaink *evmMock.ChainKeeperMock
		v      *evmMock.VoterMock
		s      *evmMock.SignerMock
		n      *evmMock.NexusMock
		msg    *types.ConfirmDepositRequest
		server types.MsgServiceServer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		basek = &evmMock.BaseKeeperMock{
			ForChainFunc: func(ctx sdk.Context, chain string) types.ChainKeeper {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return chaink
				}
				return nil
			},
		}
		chaink = &evmMock.ChainKeeperMock{
			GetDepositFunc: func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositState, bool) {
				return types.ERC20Deposit{}, 0, false
			},
			GetBurnerInfoFunc: func(sdk.Context, common.Address) *types.BurnerInfo {
				return &types.BurnerInfo{
					TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
					Symbol:       rand.StrBetween(5, 10),
					Salt:         types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				}
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) (int64, bool) { return rand.PosI64(), true },
			SetPendingDepositFunc:             func(sdk.Context, vote.PollMeta, *types.ERC20Deposit) {},
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
		}
		v = &evmMock.VoterMock{InitPollFunc: func(sdk.Context, vote.PollMeta, int64, int64) error { return nil }}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
				return rand.StrBetween(5, 20), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
		}
		chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
		n = &evmMock.NexusMock{
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}

		msg = &types.ConfirmDepositRequest{
			Sender:        rand.Bytes(20),
			Chain:         evmChain,
			TxID:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Amount:        sdk.NewUint(mathRand.Uint64()),
			BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
		}
		server = keeper.NewMsgServerImpl(basek, &evmMock.TSSMock{}, n, s, v, &mock.SnapshotterMock{})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeDepositConfirmation }), 1)
		assert.Equal(t, v.InitPollCalls()[0].Poll, chaink.SetPendingDepositCalls()[0].Poll)
	}).Repeat(repeats))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		msg.Chain = rand.StrBetween(5, 20)

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("deposit confirmed", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetDepositFunc = func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositState, bool) {
			return types.ERC20Deposit{
				TxID:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				Amount:        sdk.NewUint(mathRand.Uint64()),
				Symbol:        rand.StrBetween(5, 10),
				BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			}, types.CONFIRMED, true
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("deposit burned", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetDepositFunc = func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositState, bool) {
			return types.ERC20Deposit{
				TxID:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				Amount:        sdk.NewUint(mathRand.Uint64()),
				Symbol:        rand.StrBetween(5, 10),
				BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			}, types.BURNED, true
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("burner address unknown", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetBurnerInfoFunc = func(sdk.Context, common.Address) *types.BurnerInfo { return nil }

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitPollFunc = func(sdk.Context, vote.PollMeta, int64, int64) error { return fmt.Errorf("failed") }

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no key", testutils.Func(func(t *testing.T) {
		setup()
		s.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (string, bool) { return "", false }

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("no snapshot counter", testutils.Func(func(t *testing.T) {
		setup()
		s.GetSnapshotCounterForKeyIDFunc = func(sdk.Context, string) (int64, bool) { return 0, false }

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

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
	signedTx, err := ethTypes.SignTx(tx, ethTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
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

func newKeeper(ctx sdk.Context, chain string, confHeight int64) types.BaseKeeper {
	encCfg := app.MakeEncodingConfig()
	paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("testKey"), paramsK)
	k.SetParams(ctx, types.Params{
		Chain:               exported.Ethereum.Name,
		Network:             network,
		ConfirmationHeight:  uint64(confHeight),
		Gateway:             bytecodes,
		Token:               tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
	})
	k.ForChain(ctx, chain).SetGatewayAddress(ctx, common.HexToAddress(gateway))

	return k
}

func createMsgSignDeploy() *types.SignDeployTokenRequest {
	account := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
	symbol := rand.Str(3)
	name := rand.Str(10)
	decimals := rand.Bytes(1)[0]
	capacity := sdk.NewIntFromUint64(uint64(rand.PosI64()))

	return &types.SignDeployTokenRequest{Sender: account, TokenName: name, Symbol: symbol, Decimals: decimals, Capacity: capacity}
}
