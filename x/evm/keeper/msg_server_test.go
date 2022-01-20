package keeper_test

import (
	"fmt"
	"math/big"
	mathRand "math/rand"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	voteMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"

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
	burnerBC    = common.Hex2Bytes(types.Burnable)
	gateway     = "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7"
)

func TestCreateBurnTokens(t *testing.T) {
	var (
		evmBaseKeeper  *mock.BaseKeeperMock
		evmChainKeeper *mock.ChainKeeperMock
		tssKeeper      *mock.TSSMock
		nexusKeeper    *mock.NexusMock
		signerKeeper   *mock.SignerMock
		voteKeeper     *mock.VoterMock
		snapshotKeeper *mock.SnapshotterMock
		server         types.MsgServiceServer

		ctx            sdk.Context
		req            *types.CreateBurnTokensRequest
		secondaryKeyID tss.KeyID
	)

	repeats := 20
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		req = types.NewCreateBurnTokensRequest(rand.AccAddr(), exported.Ethereum.Name)
		secondaryKeyID = tssTestUtils.RandKeyID()

		evmChainKeeper = &mock.ChainKeeperMock{
			GetConfirmedDepositsFunc: func(ctx sdk.Context) []types.ERC20Deposit {
				return []types.ERC20Deposit{}
			},
			GetChainIDByNetworkFunc: func(ctx sdk.Context, network string) *big.Int {
				return sdk.NewIntFromBigInt(evmParams.AllCliqueProtocolChanges.ChainID).BigInt()
			},
			DeleteDepositFunc: func(ctx sdk.Context, deposit types.ERC20Deposit) {},
			SetDepositFunc:    func(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositStatus) {},
			GetBurnerInfoFunc: func(ctx sdk.Context, address types.Address) *types.BurnerInfo {
				return &types.BurnerInfo{}
			},
			EnqueueCommandFunc: func(ctx sdk.Context, cmd types.Command) error { return nil },
			GetChainIDFunc: func(sdk.Context) (*big.Int, bool) {
				return big.NewInt(rand.PosI64()), true
			},
		}
		evmBaseKeeper = &mock.BaseKeeperMock{
			ForChainFunc: func(string) types.ChainKeeper {
				return evmChainKeeper
			},
		}
		tssKeeper = &mock.TSSMock{}
		nexusKeeper = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				if chain == req.Chain {
					return exported.Ethereum, true
				}

				return nexus.Chain{}, false
			},
		}
		signerKeeper = &mock.SignerMock{
			GetNextKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return "", false
			},
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return secondaryKeyID, true
			},
		}
		voteKeeper = &mock.VoterMock{}
		snapshotKeeper = &mock.SnapshotterMock{}

		server = keeper.NewMsgServerImpl(evmBaseKeeper, tssKeeper, nexusKeeper, signerKeeper, voteKeeper, snapshotKeeper)
	}

	t.Run("should do nothing if no confirmed deposits exist", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.CreateBurnTokens(sdk.WrapSDKContext(ctx), req)

		assert.NoError(t, err)
		assert.Len(t, evmChainKeeper.DeleteDepositCalls(), 0)
	}).Repeat(repeats))

	t.Run("should return error if the next secondary key is assigned", testutils.Func(func(t *testing.T) {
		setup()

		evmChainKeeper.GetConfirmedDepositsFunc = func(ctx sdk.Context) []types.ERC20Deposit {
			return []types.ERC20Deposit{{}}
		}
		signerKeeper.GetNextKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			if chain.Name == exported.Ethereum.Name && keyRole == tss.SecondaryKey {
				return "", true
			}

			return "", false
		}

		_, err := server.CreateBurnTokens(sdk.WrapSDKContext(ctx), req)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should create burn commands", testutils.Func(func(t *testing.T) {
		setup()

		var deposits []types.ERC20Deposit
		burnerInfos := make(map[string]types.BurnerInfo)
		depositCount := int(rand.I64Between(10, 20))
		for i := 0; i < depositCount; i++ {
			deposit := types.ERC20Deposit{
				TxID:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
				Amount:           sdk.NewUint(uint64(rand.I64Between(1000, 1000000))),
				Asset:            rand.Str(5),
				DestinationChain: btc.Bitcoin.Name,
				BurnerAddress:    types.Address(common.HexToAddress(rand.HexStr(common.AddressLength))),
			}
			deposits = append(deposits, deposit)

			burnerInfos[deposit.BurnerAddress.Hex()] = types.BurnerInfo{
				TokenAddress:     types.Address(common.HexToAddress(rand.HexStr(common.AddressLength))),
				DestinationChain: deposit.DestinationChain,
				Symbol:           deposit.Asset,
				Asset:            deposit.Asset,
				Salt:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
			}
		}

		evmChainKeeper.GetConfirmedDepositsFunc = func(ctx sdk.Context) []types.ERC20Deposit {
			return deposits
		}
		evmChainKeeper.GetBurnerInfoFunc = func(ctx sdk.Context, address types.Address) *types.BurnerInfo {
			if burnerInfo, ok := burnerInfos[address.Hex()]; ok {
				return &burnerInfo
			}

			return nil
		}

		_, err := server.CreateBurnTokens(sdk.WrapSDKContext(ctx), req)

		assert.NoError(t, err)
		assert.Len(t, evmChainKeeper.DeleteDepositCalls(), depositCount)
		assert.Len(t, evmChainKeeper.SetDepositCalls(), depositCount)
		assert.Len(t, evmChainKeeper.EnqueueCommandCalls(), depositCount)

		for _, setDepositCall := range evmChainKeeper.SetDepositCalls() {
			assert.Equal(t, types.DepositStatus_Burned, setDepositCall.State)
		}

		commandIDSeen := make(map[string]bool)
		for _, command := range evmChainKeeper.EnqueueCommandCalls() {
			_, ok := commandIDSeen[command.Cmd.ID.Hex()]
			commandIDSeen[command.Cmd.ID.Hex()] = true

			assert.False(t, ok)
			assert.Equal(t, secondaryKeyID, command.Cmd.KeyID)
		}
	}).Repeat(repeats))

	t.Run("should not burn the same address multiple times", testutils.Func(func(t *testing.T) {
		setup()

		deposit1 := types.ERC20Deposit{
			TxID:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
			Amount:           sdk.NewUint(uint64(rand.I64Between(1000, 1000000))),
			Asset:            rand.Str(5),
			DestinationChain: btc.Bitcoin.Name,
			BurnerAddress:    types.Address(common.HexToAddress(rand.HexStr(common.AddressLength))),
		}
		deposit2 := types.ERC20Deposit{
			TxID:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
			Amount:           sdk.NewUint(uint64(rand.I64Between(1000, 1000000))),
			Asset:            rand.Str(5),
			DestinationChain: btc.Bitcoin.Name,
			BurnerAddress:    deposit1.BurnerAddress,
		}
		deposit3 := types.ERC20Deposit{
			TxID:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
			Amount:           sdk.NewUint(uint64(rand.I64Between(1000, 1000000))),
			Asset:            rand.Str(5),
			DestinationChain: btc.Bitcoin.Name,
			BurnerAddress:    deposit1.BurnerAddress,
		}
		burnerInfo := types.BurnerInfo{
			TokenAddress:     types.Address(common.HexToAddress(rand.HexStr(common.AddressLength))),
			DestinationChain: deposit1.DestinationChain,
			Symbol:           deposit1.Asset,
			Asset:            deposit1.Asset,
			Salt:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
		}

		evmChainKeeper.GetConfirmedDepositsFunc = func(ctx sdk.Context) []types.ERC20Deposit {
			return []types.ERC20Deposit{deposit1, deposit2, deposit3}
		}
		evmChainKeeper.GetBurnerInfoFunc = func(ctx sdk.Context, address types.Address) *types.BurnerInfo {
			return &burnerInfo
		}

		_, err := server.CreateBurnTokens(sdk.WrapSDKContext(ctx), req)

		assert.NoError(t, err)
		assert.Len(t, evmChainKeeper.DeleteDepositCalls(), 3)
		assert.Len(t, evmChainKeeper.SetDepositCalls(), 3)
		assert.Len(t, evmChainKeeper.EnqueueCommandCalls(), 1)
	}).Repeat(repeats))
}

func TestLink_UnknownChain(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encCfg := app.MakeEncodingConfig()

	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("testKey"), paramsK)
	k.ForChain(exported.Ethereum.Name).SetParams(ctx, types.Params{
		Chain:               exported.Ethereum.Name,
		Network:             network,
		ConfirmationHeight:  uint64(minConfHeight),
		GatewayCode:         bytecodes,
		TokenCode:           tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
		CommandsGasLimit:    5000000,
		TransactionFeeRate:  sdk.NewDecWithPrec(25, 5),
	})

	recipient := nexus.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	asset := rand.Str(3)

	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc:         func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false },
	}
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, &mock.SignerMock{}, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.AccAddr(), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Asset: asset})

	assert.Error(t, err)
	assert.Equal(t, 0, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 1, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_NoGateway(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encCfg := app.MakeEncodingConfig()

	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("testKey"), paramsK)
	k.ForChain(exported.Ethereum.Name).SetParams(ctx, types.Params{
		Chain:               exported.Ethereum.Name,
		Network:             network,
		ConfirmationHeight:  uint64(minConfHeight),
		GatewayCode:         bytecodes,
		TokenCode:           tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
		CommandsGasLimit:    5000000,
		TransactionFeeRate:  sdk.NewDecWithPrec(25, 5),
	})

	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	asset := rand.Str(3)

	chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}
	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, tss.KeyID) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Chain: evmChain, Sender: rand.AccAddr(), RecipientAddr: recipient.Address, Asset: asset, RecipientChain: recipient.Chain.Name})

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
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}

	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, tss.KeyID) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Chain: evmChain, Sender: rand.AccAddr(), RecipientAddr: recipient.Address, Asset: asset, RecipientChain: recipient.Chain.Name})

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
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
		IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return false },
	}

	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, tss.KeyID) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	recipient := nexus.CrossChainAddress{Address: "bcrt1q4reak3gj7xynnuc70gpeut8wxslqczhpsxhd5q8avda6m428hddqgkntss", Chain: btc.Bitcoin}
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.AccAddr(), Chain: evmChain, RecipientAddr: recipient.Address, Asset: asset, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_Success(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := rand.Context(fake.NewMultiStore())
	chain := "Ethereum"
	k := newKeeper(ctx, chain, minConfHeight)
	tokenDetails := createDetails(randomNormalizedStr(10), randomNormalizedStr(3))
	msg := createMsgSignDeploy(tokenDetails)

	k.ForChain(chain).SetPendingGateway(ctx, common.HexToAddress(gateway))
	k.ForChain(chain).ConfirmPendingGateway(ctx)

	token, err := k.ForChain(chain).CreateERC20Token(ctx, btc.Bitcoin.NativeAsset, tokenDetails, types.ZeroAddress)
	if err != nil {
		panic(err)
	}

	err = token.RecordDeployment(types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))))
	if err != nil {
		panic(err)
	}
	err = token.ConfirmDeployment()
	if err != nil {
		panic(err)
	}

	recipient := nexus.CrossChainAddress{Address: "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L", Chain: btc.Bitcoin}
	burnAddr, salt, err := k.ForChain(chain).GetBurnerAddressAndSalt(ctx, token, recipient.Address, common.HexToAddress(gateway))
	if err != nil {
		panic(err)
	}
	sender := nexus.CrossChainAddress{Address: burnAddr.Hex(), Chain: exported.Ethereum}

	chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		LinkAddressesFunc:    func(ctx sdk.Context, s nexus.CrossChainAddress, r nexus.CrossChainAddress) error { return nil },
		GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
		IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return true },
	}
	signer := &mock.SignerMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		},
		GetSnapshotCounterForKeyIDFunc: func(sdk.Context, tss.KeyID) (int64, bool) {
			return rand.PosI64(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, &mock.TSSMock{}, n, signer, &mock.VoterMock{}, &mock.SnapshotterMock{})
	_, err = server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.AccAddr(), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Asset: btc.Bitcoin.NativeAsset})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)

	expected := types.BurnerInfo{BurnerAddress: types.Address(burnAddr), TokenAddress: token.GetAddress(), DestinationChain: recipient.Chain.Name, Symbol: msg.TokenDetails.Symbol, Asset: btc.Bitcoin.NativeAsset, Salt: types.Hash(salt)}
	actual := *k.ForChain(chain).GetBurnerInfo(ctx, types.Address(burnAddr))
	assert.Equal(t, expected, actual)
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
		basek   *mock.BaseKeeperMock
		v       *mock.VoterMock
		n       *mock.NexusMock
		s       *mock.SnapshotterMock
		tssk    *mock.TSSMock
		msg     *types.ConfirmChainRequest
		voteReq *types.VoteConfirmChainRequest
		server  types.MsgServiceServer
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		chain := rand.StrBetween(5, 20)
		msg = &types.ConfirmChainRequest{
			Sender: rand.AccAddr(),
			Name:   chain,
		}
		voteReq = &types.VoteConfirmChainRequest{Name: chain}

		basek = &mock.BaseKeeperMock{
			SetPendingChainFunc: func(sdk.Context, nexus.Chain, types.Params) {},
			GetPendingChainFunc: func(_ sdk.Context, chain string) (types.PendingChain, bool) {
				if strings.EqualFold(chain, msg.Name) {
					return types.PendingChain{
						Chain:  nexus.Chain{Name: msg.Name, NativeAsset: rand.StrBetween(3, 5), SupportsForeignAssets: true, Module: rand.Str(10)},
						Params: evmTestUtils.RandomParams(),
					}, true
				}
				return types.PendingChain{}, false
			},
		}
		v = &mock.VoterMock{
			InitializePollWithSnapshotFunc: func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error {
				return nil
			},
			GetPollFunc: func(sdk.Context, vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(sdk.ValAddress, codec.ProtoMarshaler) error {
						return nil
					},
					IsFunc: func(state vote.PollState) bool {
						switch state {
						case vote.Pending:
							return true
						default:
							return false
						}
					},
				}
			},
		}

		chains := map[string]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
			IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return false },
		}
		s = &mock.SnapshotterMock{
			GetLatestSnapshotFunc: func(sdk.Context) (snapshot.Snapshot, bool) {
				return snapshot.Snapshot{Counter: rand.PosI64()}, true

			},
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return rand.ValAddr()
			},
			TakeSnapshotFunc: func(sdk.Context, tss.KeyRequirement) (snapshot.Snapshot, error) {
				return snapshot.Snapshot{Counter: rand.PosI64()}, nil
			},
		}
		tssk = &mock.TSSMock{
			GetKeyRequirementFunc: func(sdk.Context, tss.KeyRole, tss.KeyType) (tss.KeyRequirement, bool) {
				return tss.KeyRequirement{}, true
			},
		}

		server = keeper.NewMsgServerImpl(basek, tssk, n, &mock.SignerMock{}, v, s)
	}

	repeats := 20
	t.Run("happy path confirm", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeChainConfirmation }), 1)
		assert.Equal(t, 1, len(v.InitializePollWithSnapshotCalls()))
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
			GetLatestSnapshotFunc: func(sdk.Context) (snapshot.Snapshot, bool) {
				if len(s.TakeSnapshotCalls()) > 0 {
					return snapshot.Snapshot{Counter: rand.PosI64()}, true
				}
				return snapshot.Snapshot{}, false
			},
			TakeSnapshotFunc: func(sdk.Context, tss.KeyRequirement) (snapshot.Snapshot, error) {
				return snapshot.Snapshot{}, nil
			},
		}

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeChainConfirmation }), 1)
		assert.Equal(t, 1, len(v.InitializePollWithSnapshotCalls()))
	}).Repeat(repeats))

	t.Run("registered chain", testutils.Func(func(t *testing.T) {
		setup()
		msg.Name = evmChain

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		basek.GetPendingChainFunc = func(sdk.Context, string) (types.PendingChain, bool) { return types.PendingChain{}, false }

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitializePollWithSnapshotFunc = func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error {
			return fmt.Errorf("poll setup failed")
		}

		_, err := server.ConfirmChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgConfirmTokenDeploy(t *testing.T) {
	var (
		ctx     sdk.Context
		basek   *mock.BaseKeeperMock
		chaink  *mock.ChainKeeperMock
		v       *mock.VoterMock
		n       *mock.NexusMock
		s       *mock.SignerMock
		msg     *types.ConfirmTokenRequest
		token   types.ERC20Token
		server  types.MsgServiceServer
		voteReq *types.VoteConfirmTokenRequest
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		voteReq = &types.VoteConfirmTokenRequest{Chain: evmChain}
		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain string) types.ChainKeeper {
				if strings.EqualFold(chain, evmChain) {
					return chaink
				}
				return nil
			},
		}
		chaink = &mock.ChainKeeperMock{
			GetVotingThresholdFunc: func(sdk.Context) (utils.Threshold, bool) {
				return utils.Threshold{Numerator: 15, Denominator: 100}, true
			},
			GetMinVoterCountFunc: func(sdk.Context) (int64, bool) { return 15, true },
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), true
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) (int64, bool) { return rand.PosI64(), true },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
			GetERC20TokenByAssetFunc: func(ctx sdk.Context, asset string) types.ERC20Token {
				if asset == msg.Asset.Name {
					return token
				}
				return types.NilToken
			},
		}
		v = &mock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error { return nil },
			GetPollFunc: func(sdk.Context, vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(sdk.ValAddress, codec.ProtoMarshaler) error {
						return nil
					},
					IsFunc: func(state vote.PollState) bool {
						switch state {
						case vote.Pending:
							return true
						default:
							return false
						}
					},
				}
			},
		}
		chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
		n = &mock.NexusMock{
			GetChainMaintainersFunc: func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress {
				return []sdk.ValAddress{}
			},
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return tssTestUtils.RandKeyID(), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, tss.KeyID) (int64, bool) {
				return rand.PosI64(), true
			},
		}

		token = createMockERC20Token(btc.Bitcoin.NativeAsset, createDetails(randomNormalizedStr(10), randomNormalizedStr(3)))
		msg = &types.ConfirmTokenRequest{
			Sender: rand.AccAddr(),
			Chain:  evmChain,
			TxID:   types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Asset:  types.NewAsset(btc.Bitcoin.Name, btc.Bitcoin.NativeAsset),
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
		assert.Equal(t, v.InitializePollCalls()[0].Key, types.GetConfirmTokenKey(msg.TxID, btc.Bitcoin.NativeAsset))
	}).Repeat(repeats))

	t.Run("GIVEN a valid vote WHEN voting THEN event is emitted that captures vote value", testutils.Func(func(t *testing.T) {
		setup()
		hash := common.BytesToHash(rand.Bytes(common.HashLength))
		err := token.RecordDeployment(types.Hash(hash))
		if err != nil {
			panic(err)
		}
		pollKey := types.GetConfirmTokenKey(types.Hash(hash), btc.Bitcoin.NativeAsset)
		voteReq.Asset = btc.Bitcoin.NativeAsset
		voteReq.PollKey = pollKey

		_, err = server.VoteConfirmToken(sdk.WrapSDKContext(ctx), voteReq)

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

	t.Run("token unknown", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			return types.NilToken
		}

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("already registered", testutils.Func(func(t *testing.T) {
		setup()
		hash := common.BytesToHash(rand.Bytes(common.HashLength))
		token.RecordDeployment(types.Hash(hash))
		token.ConfirmDeployment()

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitializePollFunc = func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error {
			return fmt.Errorf("poll setup failed")
		}

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestAddChain(t *testing.T) {
	var (
		ctx         sdk.Context
		basek       *mock.BaseKeeperMock
		tssMock     *mock.TSSMock
		n           *mock.NexusMock
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
		basek = &mock.BaseKeeperMock{
			SetPendingChainFunc: func(sdk.Context, nexus.Chain, types.Params) {},
		}

		tssMock = &mock.TSSMock{}

		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
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
			Sender:      rand.AccAddr(),
			Name:        name,
			NativeAsset: nativeAsset,
			Params:      params,
		}

		server = keeper.NewMsgServerImpl(basek, tssMock, n, &mock.SignerMock{}, &mock.VoterMock{}, &mock.SnapshotterMock{})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(basek.SetPendingChainCalls()))
		assert.Equal(t, name, basek.SetPendingChainCalls()[0].Chain.Name)
		assert.Equal(t, nativeAsset, basek.SetPendingChainCalls()[0].Chain.NativeAsset)

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
		basek   *mock.BaseKeeperMock
		chaink  *mock.ChainKeeperMock
		v       *mock.VoterMock
		s       *mock.SignerMock
		n       *mock.NexusMock
		msg     *types.ConfirmDepositRequest
		server  types.MsgServiceServer
		voteReq *types.VoteConfirmDepositRequest
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		voteReq = &types.VoteConfirmDepositRequest{Chain: evmChain}
		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain string) types.ChainKeeper {
				if strings.EqualFold(chain, evmChain) {
					return chaink
				}
				return nil
			},
		}
		chaink = &mock.ChainKeeperMock{
			GetDepositFunc: func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, 0, false
			},
			GetBurnerInfoFunc: func(sdk.Context, types.Address) *types.BurnerInfo {
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
		v = &mock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error { return nil },
			GetPollFunc: func(sdk.Context, vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(sdk.ValAddress, codec.ProtoMarshaler) error {
						return nil
					},
					IsFunc: func(state vote.PollState) bool {
						switch state {
						case vote.Pending:
							return true
						default:
							return false
						}
					},
				}
			},
		}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return tssTestUtils.RandKeyID(), true
			},
			GetSnapshotCounterForKeyIDFunc: func(sdk.Context, tss.KeyID) (int64, bool) {
				return rand.PosI64(), true
			},
		}
		chains := map[string]nexus.Chain{
			exported.Ethereum.Name: exported.Ethereum,
			btc.Bitcoin.Name:       btc.Bitcoin,
		}
		n = &mock.NexusMock{
			GetChainMaintainersFunc: func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress { return []sdk.ValAddress{} },
			IsChainActivatedFunc:    func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}

		msg = &types.ConfirmDepositRequest{
			Sender:        rand.AccAddr(),
			Chain:         evmChain,
			TxID:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Amount:        sdk.NewUint(mathRand.Uint64()),
			BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
		}
		server = keeper.NewMsgServerImpl(basek, &mock.TSSMock{}, n, s, v, &mock.SnapshotterMock{
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
		chaink.GetDepositFunc = func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			return types.ERC20Deposit{
				TxID:             types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				Amount:           sdk.NewUint(mathRand.Uint64()),
				DestinationChain: btc.Bitcoin.Name,
				BurnerAddress:    types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			}, types.DepositStatus_Confirmed, true
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("deposit burned", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetDepositFunc = func(sdk.Context, common.Hash, common.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
			return types.ERC20Deposit{
				TxID:             types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
				Amount:           sdk.NewUint(mathRand.Uint64()),
				DestinationChain: btc.Bitcoin.Name,
				BurnerAddress:    types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			}, types.DepositStatus_Burned, true
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("burner address unknown", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetBurnerInfoFunc = func(sdk.Context, types.Address) *types.BurnerInfo { return nil }

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		v.InitializePollFunc = func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error {
			return fmt.Errorf("failed")
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgCreateDeployToken(t *testing.T) {
	var (
		ctx    sdk.Context
		basek  *mock.BaseKeeperMock
		chaink *mock.ChainKeeperMock
		v      *mock.VoterMock
		s      *mock.SignerMock
		n      *mock.NexusMock
		msg    *types.CreateDeployTokenRequest
		server types.MsgServiceServer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		msg = createMsgSignDeploy(createDetails(randomNormalizedStr(10), randomNormalizedStr(3)))

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain string) types.ChainKeeper {
				if strings.EqualFold(chain, evmChain) {
					return chaink
				}
				return nil
			},
		}
		chaink = &mock.ChainKeeperMock{
			GetParamsFunc: func(sdk.Context) types.Params {
				return types.Params{
					Chain:               exported.Ethereum.Name,
					Network:             network,
					ConfirmationHeight:  uint64(rand.I64Between(1, 10)),
					GatewayCode:         bytecodes,
					TokenCode:           tokenBC,
					Burnable:            burnerBC,
					RevoteLockingPeriod: 50,
					VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
					MinVoterCount:       15,
					CommandsGasLimit:    5000000,
				}
			},
			GetGatewayAddressFunc: func(sdk.Context) (common.Address, bool) {
				return common.BytesToAddress(rand.Bytes(common.AddressLength)), true
			},
			GetChainIDByNetworkFunc: func(ctx sdk.Context, network string) *big.Int {
				return big.NewInt(rand.I64Between(1, 1000))
			},

			CreateERC20TokenFunc: func(ctx sdk.Context, asset string, details types.TokenDetails, address types.Address) (types.ERC20Token, error) {
				if _, found := chaink.GetGatewayAddress(ctx); !found {
					return types.NilToken, fmt.Errorf("gateway address not set")
				}

				return createMockERC20Token(asset, details), nil
			},

			EnqueueCommandFunc: func(ctx sdk.Context, cmd types.Command) error { return nil },
		}

		chains := map[string]nexus.Chain{btc.Bitcoin.Name: btc.Bitcoin, exported.Ethereum.Name: exported.Ethereum}
		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
			IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return true },
			RegisterAssetFunc: func(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset) {},
		}
		s = &mock.SignerMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return tssTestUtils.RandKeyID(), true
			},
			GetNextKeyIDFunc: func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
				return "", false
			},
		}

		server = keeper.NewMsgServerImpl(basek, &mock.TSSMock{}, n, s, v, &mock.SnapshotterMock{})
	}

	repeats := 20
	t.Run("should create deploy token when gateway address is set, chains are registered and asset is registered on the origin chain ", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chaink.EnqueueCommandCalls()))
	}).Repeat(repeats))

	t.Run("should return error when chain is unknown", testutils.Func(func(t *testing.T) {
		setup()
		msg.Chain = rand.StrBetween(5, 20)

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when gateway is not set", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetGatewayAddressFunc = func(sdk.Context) (common.Address, bool) { return common.Address{}, false }

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when origin chain is unknown", testutils.Func(func(t *testing.T) {
		setup()
		msg.Asset.Chain = rand.StrBetween(5, 20)

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when asset is not registered on the origin chain", testutils.Func(func(t *testing.T) {
		setup()
		n.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return false }

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when next master key is set", testutils.Func(func(t *testing.T) {
		setup()
		s.GetNextKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return "", true
		}

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when master key is not set", testutils.Func(func(t *testing.T) {
		setup()
		s.GetCurrentKeyIDFunc = func(sdk.Context, nexus.Chain, tss.KeyRole) (tss.KeyID, bool) { return "", false }

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

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
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("testKey"), paramsK)
	k.ForChain(exported.Ethereum.Name).SetParams(ctx, types.Params{
		Chain:               exported.Ethereum.Name,
		Network:             network,
		ConfirmationHeight:  uint64(confHeight),
		GatewayCode:         bytecodes,
		TokenCode:           tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
		CommandsGasLimit:    5000000,
		Networks: []types.NetworkInfo{{
			Name: network,
			Id:   sdk.NewIntFromUint64(uint64(rand.I64Between(1, 10))),
		}},
		TransactionFeeRate: sdk.NewDecWithPrec(25, 5),
	})
	k.ForChain(chain).SetPendingGateway(ctx, common.HexToAddress(gateway))
	k.ForChain(chain).ConfirmPendingGateway(ctx)

	return k
}

func createMsgSignDeploy(details types.TokenDetails) *types.CreateDeployTokenRequest {
	account := rand.AccAddr()

	asset := types.NewAsset(btc.Bitcoin.Name, btc.Bitcoin.NativeAsset)
	return &types.CreateDeployTokenRequest{Sender: account, Chain: "Ethereum", Asset: asset, TokenDetails: details, MinAmount: sdk.NewInt(1000000)}
}

func createDetails(name, symbol string) types.TokenDetails {
	decimals := rand.Bytes(1)[0]
	capacity := sdk.NewIntFromUint64(uint64(rand.PosI64()))

	return types.NewTokenDetails(name, symbol, decimals, capacity)
}

func createMockERC20Token(asset string, details types.TokenDetails) types.ERC20Token {
	meta := types.ERC20TokenMetadata{
		Asset:        asset,
		Details:      details,
		Status:       types.Initialized,
		TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
		ChainID:      sdk.NewIntFromUint64(uint64(rand.I64Between(1, 10))),
	}
	return types.CreateERC20Token(
		func(meta types.ERC20TokenMetadata) {},
		meta,
	)
}

func randomNormalizedStr(size int) string {
	return strings.ReplaceAll(utils.NormalizeString(rand.Str(size)), utils.DefaultDelimiter, "-")
}
