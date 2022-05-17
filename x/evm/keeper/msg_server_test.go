package keeper_test

import (
	"fmt"
	"math/big"
	mathRand "math/rand"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/ethereum/go-ethereum/common"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	evmParams "github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilsMock "github.com/axelarnetwork/axelar-core/utils/mock"
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	voteMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	. "github.com/axelarnetwork/utils/test"
)

var (
	evmChain    = exported.Ethereum.Name
	network     = types.Rinkeby
	networkConf = evmParams.RinkebyChainConfig
	tokenBC     = rand.Bytes(64)
	burnerBC    = common.Hex2Bytes(types.Burnable)
	gateway     = "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7"
)

func setup() (sdk.Context, types.MsgServiceServer, *mock.BaseKeeperMock, *mock.TSSMock, *mock.NexusMock, *mock.SignerMock, *mock.VoterMock, *mock.SnapshotterMock) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

	evmBaseKeeper := &mock.BaseKeeperMock{}
	tssKeeper := &mock.TSSMock{}
	nexusKeeper := &mock.NexusMock{}
	signerKeeper := &mock.SignerMock{}
	voteKeeper := &mock.VoterMock{}
	snapshotKeeper := &mock.SnapshotterMock{}

	return ctx,
		keeper.NewMsgServerImpl(evmBaseKeeper, tssKeeper, nexusKeeper, signerKeeper, voteKeeper, snapshotKeeper),
		evmBaseKeeper, tssKeeper, nexusKeeper, signerKeeper, voteKeeper, snapshotKeeper
}

func TestSetGateway(t *testing.T) {
	req := types.NewSetGatewayRequest(rand.AccAddr(), rand.Str(5), evmTestUtils.RandomAddress())

	t.Run("should fail if any of master, secondary and external keys is not set", testutils.Func(func(t *testing.T) {
		ctx, msgServer, _, _, nexusKeeper, signerKeeper, _, _ := setup()

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			if chain == req.Chain {
				return nexus.Chain{Name: chain}, true
			}

			return nexus.Chain{}, false
		}
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return chain.Name == req.Chain }

		signerKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), keyRole != tss.MasterKey
		}
		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no master key")

		signerKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), keyRole != tss.SecondaryKey
		}
		_, err = msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no secondary key")

		signerKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		signerKeeper.GetExternalKeyIDsFunc = func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) { return []tss.KeyID{}, false }
		_, err = msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no external keys")
	}))

	t.Run("should fail if gateway is already set", testutils.Func(func(t *testing.T) {
		ctx, msgServer, baseKeeper, _, nexusKeeper, signerKeeper, _, _ := setup()
		chainKeeper := &mock.ChainKeeperMock{}

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			if chain == req.Chain {
				return nexus.Chain{Name: chain}, true
			}

			return nexus.Chain{}, false
		}
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return chain.Name == req.Chain }
		signerKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		signerKeeper.GetExternalKeyIDsFunc = func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) { return []tss.KeyID{}, true }
		baseKeeper.ForChainFunc = func(chain string) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (common.Address, bool) {
			return common.Address(evmTestUtils.RandomAddress()), true
		}

		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gateway already set")
	}))

	t.Run("should set gateway", testutils.Func(func(t *testing.T) {
		ctx, msgServer, baseKeeper, _, nexusKeeper, signerKeeper, _, _ := setup()
		chainKeeper := &mock.ChainKeeperMock{}

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
			if chain == req.Chain {
				return nexus.Chain{Name: chain}, true
			}

			return nexus.Chain{}, false
		}
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return chain.Name == req.Chain }
		signerKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.Chain, keyRole tss.KeyRole) (tss.KeyID, bool) {
			return tssTestUtils.RandKeyID(), true
		}
		signerKeeper.GetExternalKeyIDsFunc = func(ctx sdk.Context, chain nexus.Chain) ([]tss.KeyID, bool) { return []tss.KeyID{}, true }
		baseKeeper.ForChainFunc = func(chain string) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (common.Address, bool) {
			return common.Address{}, false
		}
		chainKeeper.SetGatewayFunc = func(ctx sdk.Context, address types.Address) {}

		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)
		assert.Len(t, chainKeeper.SetGatewayCalls(), 1)
		assert.Equal(t, req.Address, chainKeeper.SetGatewayCalls()[0].Address)
	}))
}

func TestSignCommands(t *testing.T) {
	setup := func() (sdk.Context, types.MsgServiceServer, *mock.BaseKeeperMock, *mock.SignerMock) {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		evmBaseKeeper := &mock.BaseKeeperMock{}
		tssKeeper := &mock.TSSMock{}
		nexusKeeper := &mock.NexusMock{}
		signerKeeper := &mock.SignerMock{}
		voteKeeper := &mock.VoterMock{}
		snapshotKeeper := &mock.SnapshotterMock{}

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain string) (nexus.Chain, bool) { return nexus.Chain{}, true }
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }

		msgServer := keeper.NewMsgServerImpl(evmBaseKeeper, tssKeeper, nexusKeeper, signerKeeper, voteKeeper, snapshotKeeper)

		return ctx, msgServer, evmBaseKeeper, signerKeeper
	}

	t.Run("should create a new command batch to sign if the latest is not being signed or aborted", testutils.Func(func(t *testing.T) {
		ctx, msgServer, evmBaseKeeper, signerKeeper := setup()

		expectedCommandIDs := make([]types.CommandID, rand.I64Between(1, 100))
		for i := range expectedCommandIDs {
			expectedCommandIDs[i] = types.NewCommandID(rand.Bytes(common.HashLength), sdk.NewInt(0))
		}
		expected := types.CommandBatchMetadata{
			ID:         rand.Bytes(common.HashLength),
			CommandIDs: expectedCommandIDs,
			Status:     types.BatchSigning,
			KeyID:      tssTestUtils.RandKeyID(),
		}

		chainKeeper := &mock.ChainKeeperMock{}
		evmBaseKeeper.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
		evmBaseKeeper.ForChainFunc = func(chain string) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(0), true }
		chainKeeper.GetLatestCommandBatchFunc = func(ctx sdk.Context) types.CommandBatch {
			return types.NonExistentCommand
		}
		chainKeeper.CreateNewBatchToSignFunc = func(ctx sdk.Context, signer types.Signer) (types.CommandBatch, error) {
			return types.NewCommandBatch(expected, func(batch types.CommandBatchMetadata) {}), nil
		}
		signerKeeper.GetSnapshotCounterForKeyIDFunc = func(ctx sdk.Context, keyID tss.KeyID) (int64, bool) { return 1, true }
		signerKeeper.StartSignFunc = func(ctx sdk.Context, info tss.SignInfo, snapshotter snapshot.Snapshotter, voter interface {
			InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
		}) error {
			return nil
		}

		res, err := msgServer.SignCommands(sdk.WrapSDKContext(ctx), types.NewSignCommandsRequest(rand.AccAddr(), rand.Str(5)))

		assert.NoError(t, err)
		assert.Equal(t, uint32(len(expected.CommandIDs)), res.CommandCount)
		assert.Equal(t, expected.ID, res.BatchedCommandsID)

		assert.Len(t, chainKeeper.CreateNewBatchToSignCalls(), 1)
		assert.Len(t, signerKeeper.StartSignCalls(), 1)
	}))

	t.Run("should get the latest if it is aborted", testutils.Func(func(t *testing.T) {
		ctx, msgServer, evmBaseKeeper, signerKeeper := setup()

		expectedCommandIDs := make([]types.CommandID, rand.I64Between(1, 100))
		for i := range expectedCommandIDs {
			expectedCommandIDs[i] = types.NewCommandID(rand.Bytes(common.HashLength), sdk.NewInt(0))
		}
		commandBatch := types.CommandBatchMetadata{
			ID:         rand.Bytes(common.HashLength),
			CommandIDs: expectedCommandIDs,
			Status:     types.BatchAborted,
			KeyID:      tssTestUtils.RandKeyID(),
		}

		chainKeeper := &mock.ChainKeeperMock{}
		evmBaseKeeper.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
		evmBaseKeeper.ForChainFunc = func(chain string) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(0), true }
		chainKeeper.GetLatestCommandBatchFunc = func(ctx sdk.Context) types.CommandBatch {
			return types.NewCommandBatch(commandBatch, func(batch types.CommandBatchMetadata) {
				assert.Equal(t, types.BatchSigning, batch.Status)
			})
		}
		signerKeeper.GetSnapshotCounterForKeyIDFunc = func(ctx sdk.Context, keyID tss.KeyID) (int64, bool) { return 1, true }
		signerKeeper.StartSignFunc = func(ctx sdk.Context, info tss.SignInfo, snapshotter snapshot.Snapshotter, voter interface {
			InitializePollWithSnapshot(ctx sdk.Context, key vote.PollKey, snapshotSeqNo int64, pollProperties ...vote.PollProperty) error
		}) error {
			return nil
		}

		res, err := msgServer.SignCommands(sdk.WrapSDKContext(ctx), types.NewSignCommandsRequest(rand.AccAddr(), rand.Str(5)))

		assert.NoError(t, err)
		assert.Equal(t, uint32(len(commandBatch.CommandIDs)), res.CommandCount)
		assert.Equal(t, commandBatch.ID, res.BatchedCommandsID)

		assert.Len(t, chainKeeper.CreateNewBatchToSignCalls(), 0)
		assert.Len(t, signerKeeper.StartSignCalls(), 1)
	}))
}

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
			GetChainIDByNetworkFunc: func(ctx sdk.Context, network string) (sdk.Int, bool) {
				return sdk.NewIntFromBigInt(evmParams.AllCliqueProtocolChanges.ChainID), true
			},
			DeleteDepositFunc: func(ctx sdk.Context, deposit types.ERC20Deposit) {},
			SetDepositFunc:    func(ctx sdk.Context, deposit types.ERC20Deposit, state types.DepositStatus) {},
			GetBurnerInfoFunc: func(ctx sdk.Context, address types.Address) *types.BurnerInfo {
				return &types.BurnerInfo{}
			},
			EnqueueCommandFunc: func(ctx sdk.Context, cmd types.Command) error { return nil },
			GetChainIDFunc: func(sdk.Context) (sdk.Int, bool) {
				return sdk.NewInt(rand.PosI64()), true
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
		evmChainKeeper.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed})
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
		evmChainKeeper.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token {
			return types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{Status: types.Confirmed})
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
		TokenCode:           tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
		CommandsGasLimit:    5000000,
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
		TokenCode:           tokenBC,
		Burnable:            burnerBC,
		RevoteLockingPeriod: 50,
		VotingThreshold:     utils.Threshold{Numerator: 15, Denominator: 100},
		MinVoterCount:       15,
		CommandsGasLimit:    5000000,
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

	k.ForChain(chain).SetGateway(ctx, types.Address(common.HexToAddress(gateway)))

	token, err := k.ForChain(chain).CreateERC20Token(ctx, btc.NativeAsset, tokenDetails, types.ZeroAddress)
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
	_, err = server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.AccAddr(), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Asset: btc.NativeAsset})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)

	expected := types.BurnerInfo{BurnerAddress: types.Address(burnAddr), TokenAddress: token.GetAddress(), DestinationChain: recipient.Chain.Name, Symbol: msg.TokenDetails.Symbol, Asset: btc.NativeAsset, Salt: types.Hash(salt)}
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

func TestHandleMsgConfirmTokenDeploy(t *testing.T) {
	var (
		ctx    sdk.Context
		basek  *mock.BaseKeeperMock
		chaink *mock.ChainKeeperMock
		v      *mock.VoterMock
		n      *mock.NexusMock
		s      *mock.SignerMock
		msg    *types.ConfirmTokenRequest
		token  types.ERC20Token
		server types.MsgServiceServer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

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
			GetParamsFunc: func(ctx sdk.Context) types.Params { return types.DefaultParams()[0] },
		}
		v = &mock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error { return nil },
			GetPollFunc: func(sdk.Context, vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (codec.ProtoMarshaler, bool, error) {
						return nil, false, nil
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

		token = createMockERC20Token(btc.NativeAsset, createDetails(randomNormalizedStr(10), randomNormalizedStr(3)))
		msg = &types.ConfirmTokenRequest{
			Sender: rand.AccAddr(),
			Chain:  evmChain,
			TxID:   types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Asset:  types.NewAsset(btc.Bitcoin.Name, btc.NativeAsset),
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
		assert.Equal(t, v.InitializePollCalls()[0].Key, types.GetConfirmTokenKey(msg.TxID, btc.NativeAsset))
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
		if err := token.RecordDeployment(types.Hash(hash)); err != nil {
			panic(err)
		}
		if err := token.ConfirmDeployment(); err != nil {
			panic(err)
		}

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
		ctx     sdk.Context
		basek   *mock.BaseKeeperMock
		chaink  *mock.ChainKeeperMock
		tssMock *mock.TSSMock
		n       *mock.NexusMock
		msg     *types.AddChainRequest
		server  types.MsgServiceServer
		name    string
		params  types.Params
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		chains := map[string]nexus.Chain{
			exported.Ethereum.Name: exported.Ethereum,
			btc.Bitcoin.Name:       btc.Bitcoin,
		}
		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(n string) types.ChainKeeper {
				if n == name {
					return chaink
				}
				return nil
			},
		}
		chaink = &mock.ChainKeeperMock{
			SetParamsFunc: func(sdk.Context, types.Params) {},
		}

		tssMock = &mock.TSSMock{}

		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain string) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
			GetChainByNativeAssetFunc: func(ctx sdk.Context, denom string) (nexus.Chain, bool) { return nexus.Chain{}, false },
			SetChainFunc:              func(sdk.Context, nexus.Chain) {},
		}

		name = rand.StrBetween(5, 20)
		params = types.DefaultParams()[0]
		params.Chain = name
		msg = &types.AddChainRequest{
			Sender: rand.AccAddr(),
			Name:   name,
			Params: params,
		}

		server = keeper.NewMsgServerImpl(basek, tssMock, n, &mock.SignerMock{}, &mock.VoterMock{}, &mock.SnapshotterMock{})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(basek.ForChainCalls()))
		assert.Equal(t, 1, len(chaink.SetParamsCalls()))
		assert.Equal(t, 1, len(n.SetChainCalls()))
		assert.Equal(t, name, basek.ForChainCalls()[0].Chain)
		assert.Equal(t, params, chaink.SetParamsCalls()[0].P)
		assert.Equal(t, name, n.SetChainCalls()[0].Chain.Name)

		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeNewChain }), 1)

	}).Repeat(repeats))

	t.Run("chain already registered", testutils.Func(func(t *testing.T) {
		setup()

		msg.Name = "Bitcoin"

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgConfirmDeposit(t *testing.T) {
	var (
		ctx    sdk.Context
		basek  *mock.BaseKeeperMock
		chaink *mock.ChainKeeperMock
		v      *mock.VoterMock
		s      *mock.SignerMock
		n      *mock.NexusMock
		msg    *types.ConfirmDepositRequest
		server types.MsgServiceServer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

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
			GetParamsFunc: func(ctx sdk.Context) types.Params { return types.DefaultParams()[0] },
		}
		v = &mock.VoterMock{
			InitializePollFunc: func(sdk.Context, vote.PollKey, []sdk.ValAddress, ...vote.PollProperty) error { return nil },
			GetPollFunc: func(sdk.Context, vote.PollKey) vote.Poll {
				return &voteMock.PollMock{
					VoteFunc: func(voter sdk.ValAddress, blockHeight int64, data codec.ProtoMarshaler) (codec.ProtoMarshaler, bool, error) {
						return nil, false, nil
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
			Sender: rand.AccAddr(),
			Chain:  evmChain,
			TxID:   types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
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
		assert.Equal(t, len(v.InitializePollCalls()), 1)
	}).Repeat(repeats))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		msg.Chain = rand.StrBetween(5, 20)

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
			GetChainIDByNetworkFunc: func(ctx sdk.Context, network string) (sdk.Int, bool) {
				return sdk.NewInt(rand.I64Between(1, 1000)), true
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
			RegisterAssetFunc:     func(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset) error { return nil },
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

func TestRetryFailedEvent(t *testing.T) {
	var (
		ctx sdk.Context
		bk  *mock.BaseKeeperMock
		ck  *mock.ChainKeeperMock
		n   *mock.NexusMock
	)

	ctx, msgServer, bk, _, n, _, _, _ := setup()
	contractCallQueue := &utilsMock.KVQueueMock{
		EnqueueFunc: func(key utils.Key, value codec.ProtoMarshaler) {},
	}
	ck = &mock.ChainKeeperMock{
		GetConfirmedEventQueueFunc: func(ctx sdk.Context) utils.KVQueue {
			return contractCallQueue
		},
	}
	bk.ForChainFunc = func(chain string) types.ChainKeeper {
		return ck
	}
	bk.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }

	req := types.NewRetryFailedEventRequest(rand.AccAddr(), rand.Str(5), rand.Str(5))

	chainFound := func(found bool) func() {
		return func() {
			n.GetChainFunc = func(sdk.Context, string) (nexus.Chain, bool) {
				if !found {
					return nexus.Chain{}, false
				}
				return nexus.Chain{}, true
			}
		}
	}

	isChainActivated := func(isActivated bool) func() {
		return func() {
			n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool {
				return isActivated
			}
		}
	}

	eventFound := func(found bool, eventStatus types.Event_Status) func() {
		return func() {
			ck.GetEventFunc = func(sdk.Context, types.EventID) (types.Event, bool) {
				if !found {
					return types.Event{}, false
				}
				return types.Event{Status: eventStatus}, true
			}
		}
	}

	When("chain is not found", chainFound(false)).
		Then("should return error", func(t *testing.T) {
			_, err := msgServer.RetryFailedEvent(sdk.WrapSDKContext(ctx), req)
			assert.Error(t, err)
		}).
		Run(t)

	When("chain is found", chainFound(true)).
		When("chain is not activated", isChainActivated(false)).
		Then("should return error", func(t *testing.T) {
			_, err := msgServer.RetryFailedEvent(sdk.WrapSDKContext(ctx), req)
			assert.Error(t, err)
		}).
		Run(t)

	When("chain is found", chainFound(true)).
		When("chain is activated", isChainActivated(true)).
		When("event not found", eventFound(false, types.EventNonExistent)).
		Then("should return error", func(t *testing.T) {
			_, err := msgServer.RetryFailedEvent(sdk.WrapSDKContext(ctx), req)
			assert.Error(t, err)
		}).
		Run(t)

	When("chain is found", chainFound(true)).
		When("chain is activated", isChainActivated(true)).
		When("event is completed", eventFound(true, types.EventCompleted)).
		Then("should return error", func(t *testing.T) {
			_, err := msgServer.RetryFailedEvent(sdk.WrapSDKContext(ctx), req)
			assert.Error(t, err)
		}).
		Run(t)

	When("chain is found", chainFound(true)).
		When("chain is activated", isChainActivated(true)).
		When("event is failed", eventFound(true, types.EventFailed)).
		Then("should retry event", func(t *testing.T) {
			_, err := msgServer.RetryFailedEvent(sdk.WrapSDKContext(ctx), req)
			assert.NoError(t, err)
			assert.Len(t, contractCallQueue.EnqueueCalls(), 1)
		}).
		Run(t)
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
	})
	k.ForChain(chain).SetGateway(ctx, types.Address(common.HexToAddress(gateway)))

	return k
}

func createMsgSignDeploy(details types.TokenDetails) *types.CreateDeployTokenRequest {
	account := rand.AccAddr()

	asset := types.NewAsset(btc.Bitcoin.Name, btc.NativeAsset)
	return &types.CreateDeployTokenRequest{Sender: account, Chain: "Ethereum", Asset: asset, TokenDetails: details}
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
