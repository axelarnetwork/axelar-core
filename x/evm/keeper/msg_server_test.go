package keeper_test

import (
	"fmt"
	voteMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/cosmos/cosmos-sdk/codec"
	"math/big"
	mathRand "math/rand"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	evmTypes "github.com/ethereum/go-ethereum/core/types"
	evmParams "github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmMock "github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	evmChain    = exported.Ethereum.Name
	network     = types.Rinkeby
	networkConf = evmParams.RinkebyChainConfig
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
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
	})

	recipient := nexus.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	asset := rand.Str(3)

	n := &evmMock.NexusMock{
		GetChainFunc: func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false },
	}
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, &mock.SignerMock{}, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.Bytes(sdk.AddrLen), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Asset: asset})

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
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
	})

	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	asset := rand.Str(3)

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
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Chain: evmChain, Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, Asset: asset, RecipientChain: recipient.Chain.Name})

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
	asset := rand.Str(3)

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
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Chain: evmChain, Sender: rand.Bytes(sdk.AddrLen), RecipientAddr: recipient.Address, Asset: asset, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 0, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_NoRegisteredAsset(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := newKeeper(ctx, "Ethereum", minConfHeight)

	asset := rand.Str(3)

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
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.Bytes(sdk.AddrLen), Chain: evmChain, RecipientAddr: recipient.Address, Asset: asset, RecipientChain: recipient.Chain.Name})

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

	k.ForChain(ctx, chain).SetTokenInfo(ctx, btc.Bitcoin.NativeAsset, msg)

	recipient := nexus.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	tokenAddr, err := k.ForChain(ctx, chain).GetTokenAddress(ctx, btc.Bitcoin.NativeAsset, common.HexToAddress(gateway))
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
	_, err = server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.Bytes(sdk.AddrLen), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Asset: btc.Bitcoin.NativeAsset})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)

	assert.Equal(t, types.BurnerInfo{TokenAddress: types.Address(tokenAddr), DestinationChain: recipient.Chain.Name, Symbol: msg.Symbol, Asset: btc.Bitcoin.NativeAsset, Salt: types.Hash(salt)}, *k.ForChain(ctx, chain).GetBurnerInfo(ctx, burnAddr))
}

func TestDeployTx_DifferentValue_DifferentHash(t *testing.T) {
	tx1 := createSignedDeployTx()
	privateKey, err := evmCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = evmTypes.SignTx(tx1, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newValue := big.NewInt(rand.I64Between(1, 10000))
	tx2 := sign(evmTypes.NewContractCreation(tx1.Nonce(), newValue, tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = evmTypes.SignTx(tx2, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestDeployTx_DifferentData_DifferentHash(t *testing.T) {
	tx1 := createSignedDeployTx()
	privateKey, err := evmCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = evmTypes.SignTx(tx1, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newData := rand.Bytes(int(rand.I64Between(1, 10000)))
	tx2 := sign(evmTypes.NewContractCreation(tx1.Nonce(), tx1.Value(), tx1.Gas(), tx1.GasPrice(), newData))
	tx2, err = evmTypes.SignTx(tx2, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestMintTx_DifferentValue_DifferentHash(t *testing.T) {
	tx1 := createSignedTx()
	privateKey, err := evmCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = evmTypes.SignTx(tx1, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newValue := big.NewInt(rand.I64Between(1, 10000))
	tx2 := sign(evmTypes.NewTransaction(tx1.Nonce(), *tx1.To(), newValue, tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = evmTypes.SignTx(tx2, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestMintTx_DifferentData_DifferentHash(t *testing.T) {
	tx1 := createSignedTx()
	privateKey, err := evmCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = evmTypes.SignTx(tx1, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newData := rand.Bytes(int(rand.I64Between(1, 10000)))
	tx2 := sign(evmTypes.NewTransaction(tx1.Nonce(), *tx1.To(), tx1.Value(), tx1.Gas(), tx1.GasPrice(), newData))
	tx2, err = evmTypes.SignTx(tx2, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestMintTx_DifferentRecipient_DifferentHash(t *testing.T) {
	tx1 := createSignedTx()
	privateKey, err := evmCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	tx1, err = evmTypes.SignTx(tx1, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	newTo := common.BytesToAddress(rand.Bytes(common.AddressLength))
	tx2 := sign(evmTypes.NewTransaction(tx1.Nonce(), newTo, tx1.Value(), tx1.Gas(), tx1.GasPrice(), tx1.Data()))
	tx2, err = evmTypes.SignTx(tx2, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	assert.NotEqual(t, tx1.Hash(), tx2.Hash())
}

func TestHandleMsgConfirmChain(t *testing.T) {
	var (
		ctx     sdk.Context
		basek   *evmMock.BaseKeeperMock
		chaink  *evmMock.ChainKeeperMock
		v       *evmMock.VoterMock
		n       *evmMock.NexusMock
		s       *evmMock.SnapshotterMock
		msg     *types.ConfirmChainRequest
		voteReq *types.VoteConfirmChainRequest
		server  types.MsgServiceServer
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		chain := rand.StrBetween(5, 20)
		msg = &types.ConfirmChainRequest{
			Sender: rand.Bytes(20),
			Name:   chain,
		}
		voteReq = &types.VoteConfirmChainRequest{Name: chain}

		chaink = &evmMock.ChainKeeperMock{
			GetRevoteLockingPeriodFunc: func(ctx sdk.Context) (int64, bool) {
				return rand.I64Between(50, 100), true
			},
			GetVotingThresholdFunc: func(sdk.Context) (utils.Threshold, bool) {
				return utils.Threshold{Numerator: 15, Denominator: 100}, true
			},
			GetMinVoterCountFunc: func(sdk.Context) (int64, bool) { return 15, true },
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
		v = &evmMock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error {
				return nil
			},
			GetPollFunc: func(ctx2 sdk.Context, key vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(sdk.ValAddress, codec.ProtoMarshaler) error {
						return nil
					},
					IsFunc: func(vote.PollState) bool {
						return true
					},
				}
			},
		}

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
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return rand.ValAddr()
			},
		}

		server = keeper.NewMsgServerImpl(basek, &mock.TSSMock{}, n, &mock.SignerMock{}, v, s)
	}

	repeats := 20
	t.Run("happy path confirm", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeChainConfirmation }), 1)
		assert.Equal(t, 1, len(v.InitializePollCalls()))
	}).Repeat(repeats))

	t.Run("GIVEN a valid vote WHEN voting THEN event is emitted that captures vote value", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.VoteConfirmChain(sdk.WrapSDKContext(ctx), voteReq)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeChainConfirmation
			if !isValidType {
				return false
			}
			isVoteAction := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				return string(attribute.GetKey()) == sdk.AttributeKeyAction &&
					string(attribute.GetValue()) == types.AttributeValueVote
			})) == 1
			if !isVoteAction {
				return false
			}
			hasCorrectValue := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				if string(attribute.GetKey()) != types.AttributeKeyValue {
					return false
				}
				return string(attribute.GetValue()) == "false"
			})) == 1
			return hasCorrectValue
		}), 1)

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
			TakeSnapshotFunc: func(sdk.Context, tss.KeyRequirement) (snapshot.Snapshot, error) {
				return snapshot.Snapshot{}, nil
			},
		}

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeChainConfirmation }), 1)
		assert.Equal(t, 1, len(v.InitializePollCalls()))
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
		v.InitializePollFunc = func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error {
			return fmt.Errorf("poll setup failed")
		}

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgConfirmTokenDeploy(t *testing.T) {
	var (
		ctx     sdk.Context
		basek   *evmMock.BaseKeeperMock
		chaink  *evmMock.ChainKeeperMock
		v       *evmMock.VoterMock
		n       *evmMock.NexusMock
		s       *evmMock.SignerMock
		msg     *types.ConfirmTokenRequest
		server  types.MsgServiceServer
		voteReq *types.VoteConfirmTokenRequest
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		voteReq = &types.VoteConfirmTokenRequest{Chain: evmChain}
		basek = &evmMock.BaseKeeperMock{
			ForChainFunc: func(ctx sdk.Context, chain string) types.ChainKeeper {
				if strings.ToLower(chain) == strings.ToLower(evmChain) {
					return chaink
				}
				return nil
			},
		}
		chaink = &evmMock.ChainKeeperMock{
			GetVotingThresholdFunc: func(sdk.Context) (utils.Threshold, bool) {
				return utils.Threshold{Numerator: 15, Denominator: 100}, true
			},
			GetMinVoterCountFunc: func(sdk.Context) (int64, bool) { return 15, true },
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), true
			},
			GetTokenAddressFunc: func(sdk.Context, string, common.Address) (common.Address, error) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), nil
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) (int64, bool) { return rand.PosI64(), true },
			GetTokenSymbolFunc:                func(sdk.Context, string) (string, bool) { return rand.StrBetween(3, 5), true },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
			SetPendingTokenDeploymentFunc:     func(sdk.Context, vote.PollKey, types.ERC20TokenDeployment) {},
			GetPendingTokenDeploymentFunc: func(sdk.Context, vote.PollKey) (types.ERC20TokenDeployment, bool) {
				return types.ERC20TokenDeployment{Asset: ""}, true
			},
		}
		v = &evmMock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error { return nil },
			GetPollFunc: func(ctx2 sdk.Context, key vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(sdk.ValAddress, codec.ProtoMarshaler) error {
						return nil
					},
					IsFunc: func(vote.PollState) bool {
						return true
					},
				}
			},
		}
		chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
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
			Sender:      rand.Bytes(20),
			Chain:       evmChain,
			TxID:        types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			OriginChain: btc.Bitcoin.Name,
		}

		server = keeper.NewMsgServerImpl(basek, &mock.TSSMock{}, n, s, v, &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return rand.ValAddr()
			}})
	}

	repeats := 20
	t.Run("happy path confirm", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeTokenConfirmation }), 1)
		assert.Equal(t, v.InitializePollCalls()[0].Key, chaink.SetPendingTokenDeploymentCalls()[0].PollKey)
	}).Repeat(repeats))

	t.Run("GIVEN a valid vote WHEN voting THEN event is emitted that captures vote value", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.VoteConfirmToken(sdk.WrapSDKContext(ctx), voteReq)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeTokenConfirmation
			if !isValidType {
				return false
			}
			isVoteAction := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				return string(attribute.GetKey()) == sdk.AttributeKeyAction &&
					string(attribute.GetValue()) == types.AttributeValueVote
			})) == 1
			if !isVoteAction {
				return false
			}
			hasCorrectValue := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				if string(attribute.GetKey()) != types.AttributeKeyValue {
					return false
				}
				return string(attribute.GetValue()) == "false"
			})) == 1
			return hasCorrectValue
		}), 1)
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
		v.InitializePollFunc = func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error {
			return fmt.Errorf("poll setup failed")
		}

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

		tssMock = &evmMock.TSSMock{}

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
		msg = &types.AddChainRequest{
			Sender:      rand.Bytes(20),
			Name:        name,
			NativeAsset: nativeAsset,
			Params:      params,
		}

		server = keeper.NewMsgServerImpl(basek, tssMock, n, &evmMock.SignerMock{}, &evmMock.VoterMock{}, &mock.SnapshotterMock{})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(basek.SetParamsCalls()))
		assert.Equal(t, 1, len(basek.SetPendingChainCalls()))
		assert.Equal(t, name, basek.SetPendingChainCalls()[0].Chain.Name)
		assert.Equal(t, nativeAsset, basek.SetPendingChainCalls()[0].Chain.NativeAsset)
		assert.Equal(t, name, basek.SetParamsCalls()[0].Params[0].Chain)

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
		ctx     sdk.Context
		basek   *evmMock.BaseKeeperMock
		chaink  *evmMock.ChainKeeperMock
		v       *evmMock.VoterMock
		s       *evmMock.SignerMock
		n       *evmMock.NexusMock
		msg     *types.ConfirmDepositRequest
		server  types.MsgServiceServer
		voteReq *types.VoteConfirmDepositRequest
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		voteReq = &types.VoteConfirmDepositRequest{Chain: evmChain}
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
			SetPendingDepositFunc:             func(sdk.Context, vote.PollKey, *types.ERC20Deposit) {},
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
			GetVotingThresholdFunc: func(sdk.Context) (utils.Threshold, bool) {
				return utils.Threshold{Numerator: 15, Denominator: 100}, true
			},
			GetMinVoterCountFunc: func(sdk.Context) (int64, bool) { return 15, true },
			GetPendingDepositFunc: func(sdk.Context, vote.PollKey) (types.ERC20Deposit, bool) {
				return types.ERC20Deposit{
					DestinationChain: evmChain,
				}, true
			},
		}
		v = &evmMock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error { return nil },
			GetPollFunc: func(ctx2 sdk.Context, key vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(sdk.ValAddress, codec.ProtoMarshaler) error {
						return nil
					},
					IsFunc: func(vote.PollState) bool {
						return true
					},
				}
			},
		}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (string, bool) {
				return rand.StrBetween(5, 20), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, string) (int64, bool) {
				return rand.PosI64(), true
			},
		}
		chains := map[string]nexus.Chain{
			exported.Ethereum.Name: exported.Ethereum,
			btc.Bitcoin.Name:       btc.Bitcoin,
		}
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
		server = keeper.NewMsgServerImpl(basek, &evmMock.TSSMock{}, n, s, v, &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return rand.ValAddr()
			},
		})
	}

	repeats := 20
	t.Run("happy path confirm", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeDepositConfirmation }), 1)
		assert.Equal(t, v.InitializePollCalls()[0].Key, chaink.SetPendingDepositCalls()[0].Key)
	}).Repeat(repeats))

	t.Run("GIVEN a valid vote WHEN voting THEN event is emitted that captures vote value", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.VoteConfirmDeposit(sdk.WrapSDKContext(ctx), voteReq)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool {
			isValidType := event.GetType() == types.EventTypeDepositConfirmation
			if !isValidType {
				return false
			}
			isVoteAction := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				return string(attribute.GetKey()) == sdk.AttributeKeyAction &&
					string(attribute.GetValue()) == types.AttributeValueVote
			})) == 1
			if !isVoteAction {
				return false
			}
			hasCorrectValue := len(testutils.Attributes(event.GetAttributes()).Filter(func(attribute abci.EventAttribute) bool {
				if string(attribute.GetKey()) != types.AttributeKeyValue {
					return false
				}
				return string(attribute.GetValue()) == "false"
			})) == 1
			return hasCorrectValue
		}), 1)

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
				TxID:             types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				Amount:           sdk.NewUint(mathRand.Uint64()),
				DestinationChain: btc.Bitcoin.Name,
				BurnerAddress:    types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			}, types.CONFIRMED, true
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("deposit burned", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetDepositFunc = func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositState, bool) {
			return types.ERC20Deposit{
				TxID:             types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				Amount:           sdk.NewUint(mathRand.Uint64()),
				DestinationChain: btc.Bitcoin.Name,
				BurnerAddress:    types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
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
		v.InitializePollFunc = func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error { return fmt.Errorf("failed") }

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

func createSignedDeployTx() *evmTypes.Transaction {
	generator := rand.PInt64Gen()

	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(generator.Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)
	byteCode := rand.Bytes(int(rand.I64Between(1, 10000)))

	return sign(evmTypes.NewContractCreation(nonce, value, gasLimit, gasPrice, byteCode))
}

func sign(tx *evmTypes.Transaction) *evmTypes.Transaction {
	privateKey, err := evmCrypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	signedTx, err := evmTypes.SignTx(tx, evmTypes.NewEIP155Signer(networkConf.ChainID), privateKey)
	if err != nil {
		panic(err)
	}
	return signedTx
}

func createSignedTx() *evmTypes.Transaction {
	generator := rand.PInt64Gen()
	contractAddr := common.BytesToAddress(rand.Bytes(common.AddressLength))
	nonce := uint64(generator.Next())
	gasPrice := big.NewInt(rand.PInt64Gen().Next())
	gasLimit := uint64(generator.Next())
	value := big.NewInt(0)

	data := rand.Bytes(int(rand.I64Between(0, 1000)))
	return sign(evmTypes.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data))
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
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
	})
	k.ForChain(ctx, chain).SetGatewayAddress(ctx, common.HexToAddress(gateway))

	return k
}

func createMsgSignDeploy() *types.CreateDeployTokenRequest {
	account := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
	symbol := rand.Str(3)
	name := rand.Str(10)
	decimals := rand.Bytes(1)[0]
	capacity := sdk.NewIntFromUint64(uint64(rand.PosI64()))

	return &types.CreateDeployTokenRequest{Sender: account, TokenName: name, Symbol: symbol, Decimals: decimals, Capacity: capacity}
}
