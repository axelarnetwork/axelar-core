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
	"github.com/ethereum/go-ethereum/crypto"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	evmParams "github.com/ethereum/go-ethereum/params"
	"github.com/gogo/protobuf/proto"
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
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTestUtils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/funcs"
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

func setup() (sdk.Context, types.MsgServiceServer, *mock.BaseKeeperMock, *mock.NexusMock, *mock.VoterMock, *mock.SnapshotterMock, *mock.MultisigKeeperMock) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

	evmBaseKeeper := &mock.BaseKeeperMock{}
	nexusKeeper := &mock.NexusMock{}
	voteKeeper := &mock.VoterMock{}
	snapshotKeeper := &mock.SnapshotterMock{}
	stakingKeeper := &mock.StakingKeeperMock{}
	slashingKeeper := &mock.SlashingKeeperMock{}
	multisigKeeper := &mock.MultisigKeeperMock{}

	return ctx,
		keeper.NewMsgServerImpl(evmBaseKeeper, nexusKeeper, voteKeeper, snapshotKeeper, stakingKeeper, slashingKeeper, multisigKeeper),
		evmBaseKeeper, nexusKeeper, voteKeeper, snapshotKeeper, multisigKeeper
}

func TestSetGateway(t *testing.T) {
	req := types.NewSetGatewayRequest(rand.AccAddr(), rand.Str(5), evmTestUtils.RandomAddress())

	t.Run("should fail if current key is not set", testutils.Func(func(t *testing.T) {
		ctx, msgServer, _, nexusKeeper, _, _, multisigKeeper := setup()

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			if chain == req.Chain {
				return nexus.Chain{Name: chain}, true
			}

			return nexus.Chain{}, false
		}
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return chain.Name == req.Chain }

		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
			return "", false
		}
		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "current key not set for chain")
	}))

	t.Run("should fail if gateway is already set", testutils.Func(func(t *testing.T) {
		ctx, msgServer, baseKeeper, nexusKeeper, _, _, multisigKeeper := setup()
		chainKeeper := &mock.ChainKeeperMock{}

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			if chain == req.Chain {
				return nexus.Chain{Name: chain}, true
			}

			return nexus.Chain{}, false
		}
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return chain.Name == req.Chain }
		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		}
		baseKeeper.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) {
			return evmTestUtils.RandomAddress(), true
		}

		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gateway already set")
	}))

	t.Run("should set gateway", testutils.Func(func(t *testing.T) {
		ctx, msgServer, baseKeeper, nexusKeeper, _, _, multisigKeeper := setup()
		chainKeeper := &mock.ChainKeeperMock{}

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			if chain == req.Chain {
				return nexus.Chain{Name: chain}, true
			}

			return nexus.Chain{}, false
		}
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return chain.Name == req.Chain }
		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		}
		baseKeeper.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) {
			return types.Address{}, false
		}
		chainKeeper.SetGatewayFunc = func(ctx sdk.Context, address types.Address) {}

		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)
		assert.Len(t, chainKeeper.SetGatewayCalls(), 1)
		assert.Equal(t, req.Address, chainKeeper.SetGatewayCalls()[0].Address)
	}))
}

func TestSignCommands(t *testing.T) {
	setup := func() (sdk.Context, types.MsgServiceServer, *mock.BaseKeeperMock, *mock.MultisigKeeperMock) {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())

		evmBaseKeeper := &mock.BaseKeeperMock{}
		nexusKeeper := &mock.NexusMock{}
		voteKeeper := &mock.VoterMock{}
		snapshotKeeper := &mock.SnapshotterMock{}
		stakingKeeper := &mock.StakingKeeperMock{}
		slashingKeeper := &mock.SlashingKeeperMock{}
		multisigKeeper := &mock.MultisigKeeperMock{}

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, true }
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }

		msgServer := keeper.NewMsgServerImpl(evmBaseKeeper, nexusKeeper, voteKeeper, snapshotKeeper, stakingKeeper, slashingKeeper, multisigKeeper)

		return ctx, msgServer, evmBaseKeeper, multisigKeeper
	}

	t.Run("should create a new command batch to sign if the latest is not being signed or aborted", testutils.Func(func(t *testing.T) {
		ctx, msgServer, evmBaseKeeper, multisigKeeper := setup()

		expectedCommandIDs := make([]types.CommandID, rand.I64Between(1, 100))
		for i := range expectedCommandIDs {
			expectedCommandIDs[i] = types.NewCommandID(rand.Bytes(common.HashLength), sdk.NewInt(0))
		}
		expected := types.CommandBatchMetadata{
			ID:         rand.Bytes(common.HashLength),
			CommandIDs: expectedCommandIDs,
			Status:     types.BatchSigning,
			KeyID:      multisigTestUtils.KeyID(),
		}

		chainKeeper := &mock.ChainKeeperMock{}
		evmBaseKeeper.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
		evmBaseKeeper.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(0), true }
		chainKeeper.GetLatestCommandBatchFunc = func(ctx sdk.Context) types.CommandBatch {
			return types.NonExistentCommand
		}
		chainKeeper.CreateNewBatchToSignFunc = func(ctx sdk.Context) (types.CommandBatch, error) {
			return types.NewCommandBatch(expected, func(batch types.CommandBatchMetadata) {}), nil
		}
		multisigKeeper.SignFunc = func(ctx sdk.Context, keyID multisig.KeyID, payloadHash multisig.Hash, module string, moduleMetadata ...codec.ProtoMarshaler) error {
			return nil
		}

		res, err := msgServer.SignCommands(sdk.WrapSDKContext(ctx), types.NewSignCommandsRequest(rand.AccAddr(), rand.Str(5)))

		assert.NoError(t, err)
		assert.Equal(t, uint32(len(expected.CommandIDs)), res.CommandCount)
		assert.Equal(t, expected.ID, res.BatchedCommandsID)

		assert.Len(t, chainKeeper.CreateNewBatchToSignCalls(), 1)
		assert.Len(t, multisigKeeper.SignCalls(), 1)
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
			KeyID:      multisigTestUtils.KeyID(),
		}

		chainKeeper := &mock.ChainKeeperMock{}
		evmBaseKeeper.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
		evmBaseKeeper.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper { return chainKeeper }
		chainKeeper.GetChainIDFunc = func(ctx sdk.Context) (sdk.Int, bool) { return sdk.NewInt(0), true }
		chainKeeper.GetLatestCommandBatchFunc = func(ctx sdk.Context) types.CommandBatch {
			return types.NewCommandBatch(commandBatch, func(batch types.CommandBatchMetadata) {
				assert.Equal(t, types.BatchSigning, batch.Status)
			})
		}
		signerKeeper.SignFunc = func(ctx sdk.Context, keyID multisig.KeyID, payloadHash multisig.Hash, module string, moduleMetadata ...codec.ProtoMarshaler) error {
			return nil
		}

		res, err := msgServer.SignCommands(sdk.WrapSDKContext(ctx), types.NewSignCommandsRequest(rand.AccAddr(), rand.Str(5)))

		assert.NoError(t, err)
		assert.Equal(t, uint32(len(commandBatch.CommandIDs)), res.CommandCount)
		assert.Equal(t, commandBatch.ID, res.BatchedCommandsID)

		assert.Len(t, chainKeeper.CreateNewBatchToSignCalls(), 0)
		assert.Len(t, signerKeeper.SignCalls(), 1)
	}))
}

func TestCreateBurnTokens(t *testing.T) {
	var (
		evmBaseKeeper  *mock.BaseKeeperMock
		evmChainKeeper *mock.ChainKeeperMock
		nexusKeeper    *mock.NexusMock
		voteKeeper     *mock.VoterMock
		multisigKeeper *mock.MultisigKeeperMock
		snapshotKeeper *mock.SnapshotterMock
		server         types.MsgServiceServer

		ctx   sdk.Context
		req   *types.CreateBurnTokensRequest
		keyID multisig.KeyID
	)

	repeats := 20
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		req = types.NewCreateBurnTokensRequest(rand.AccAddr(), exported.Ethereum.Name.String())
		keyID = multisigTestUtils.KeyID()

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
			ForChainFunc: func(nexus.ChainName) types.ChainKeeper {
				return evmChainKeeper
			},
		}
		nexusKeeper = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				if chain == req.Chain {
					return exported.Ethereum, true
				}

				return nexus.Chain{}, false
			},
		}
		multisigKeeper = &mock.MultisigKeeperMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
				return keyID, true
			},
		}
		voteKeeper = &mock.VoterMock{}
		snapshotKeeper = &mock.SnapshotterMock{}
		stakingKeeper := &mock.StakingKeeperMock{}
		slashingKeeper := &mock.SlashingKeeperMock{}

		server = keeper.NewMsgServerImpl(evmBaseKeeper, nexusKeeper, voteKeeper, snapshotKeeper, stakingKeeper, slashingKeeper, multisigKeeper)
	}

	t.Run("should do nothing if no confirmed deposits exist", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.CreateBurnTokens(sdk.WrapSDKContext(ctx), req)

		assert.NoError(t, err)
		assert.Len(t, evmChainKeeper.DeleteDepositCalls(), 0)
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
				DestinationChain: axelarnet.Axelarnet.Name,
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
			assert.EqualValues(t, keyID, command.Cmd.KeyID)
		}
	}).Repeat(repeats))

	t.Run("should not burn the same address multiple times", testutils.Func(func(t *testing.T) {
		setup()

		deposit1 := types.ERC20Deposit{
			TxID:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
			Amount:           sdk.NewUint(uint64(rand.I64Between(1000, 1000000))),
			Asset:            rand.Str(5),
			DestinationChain: axelarnet.Axelarnet.Name,
			BurnerAddress:    types.Address(common.HexToAddress(rand.HexStr(common.AddressLength))),
		}
		deposit2 := types.ERC20Deposit{
			TxID:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
			Amount:           sdk.NewUint(uint64(rand.I64Between(1000, 1000000))),
			Asset:            rand.Str(5),
			DestinationChain: axelarnet.Axelarnet.Name,
			BurnerAddress:    deposit1.BurnerAddress,
		}
		deposit3 := types.ERC20Deposit{
			TxID:             types.Hash(common.HexToHash(rand.HexStr(common.HashLength))),
			Amount:           sdk.NewUint(uint64(rand.I64Between(1000, 1000000))),
			Asset:            rand.Str(5),
			DestinationChain: axelarnet.Axelarnet.Name,
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
		EndBlockerLimit:     50,
	})

	recipient := nexus.CrossChainAddress{Address: rand.ValAddr().String(), Chain: axelarnet.Axelarnet}
	asset := rand.Str(3)

	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc:         func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, false },
	}
	server := keeper.NewMsgServerImpl(k, n, &mock.VoterMock{}, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, &mock.MultisigKeeperMock{})
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
		EndBlockerLimit:     50,
	})

	recipient := nexus.CrossChainAddress{Address: rand.ValAddr().String(), Chain: axelarnet.Axelarnet}
	asset := rand.Str(3)

	chains := map[nexus.ChainName]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}
	multisigKeeper := &mock.MultisigKeeperMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, n, &mock.VoterMock{}, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, multisigKeeper)
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

	recipient := nexus.CrossChainAddress{Address: rand.ValAddr().String(), Chain: axelarnet.Axelarnet}
	asset := rand.Str(3)

	chains := map[nexus.ChainName]nexus.Chain{exported.Ethereum.Name: exported.Ethereum}
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
	}

	multisigKeeper := &mock.MultisigKeeperMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, n, &mock.VoterMock{}, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, multisigKeeper)
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

	chains := map[nexus.ChainName]nexus.Chain{axelarnet.Axelarnet.Name: axelarnet.Axelarnet, exported.Ethereum.Name: exported.Ethereum}
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
		IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return false },
	}

	multisigKeeper := &mock.MultisigKeeperMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, n, &mock.VoterMock{}, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, multisigKeeper)
	recipient := nexus.CrossChainAddress{Address: rand.ValAddr().String(), Chain: axelarnet.Axelarnet}
	_, err := server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.AccAddr(), Chain: evmChain, RecipientAddr: recipient.Address, Asset: asset, RecipientChain: recipient.Chain.Name})

	assert.Error(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 0, len(n.LinkAddressesCalls()))
}

func TestLink_Success(t *testing.T) {
	minConfHeight := rand.I64Between(1, 10)
	ctx := rand.Context(fake.NewMultiStore())
	chain := nexus.ChainName("Ethereum")
	k := newKeeper(ctx, chain, minConfHeight)
	tokenDetails := createDetails(randomNormalizedStr(10), randomNormalizedStr(3))
	msg := createMsgSignDeploy(tokenDetails)

	k.ForChain(chain).SetGateway(ctx, types.Address(common.HexToAddress(gateway)))

	token, err := k.ForChain(chain).CreateERC20Token(ctx, axelarnet.NativeAsset, tokenDetails, types.ZeroAddress)
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

	recipient := nexus.CrossChainAddress{Address: rand.ValAddr().String(), Chain: axelarnet.Axelarnet}
	salt := k.ForChain(chain).GenerateSalt(ctx, recipient.Address)
	burnAddr, err := k.ForChain(chain).GetBurnerAddress(ctx, token, salt, types.Address(common.HexToAddress(gateway)))
	if err != nil {
		panic(err)
	}
	sender := nexus.CrossChainAddress{Address: burnAddr.Hex(), Chain: exported.Ethereum}

	chains := map[nexus.ChainName]nexus.Chain{axelarnet.Axelarnet.Name: axelarnet.Axelarnet, exported.Ethereum.Name: exported.Ethereum}
	n := &mock.NexusMock{
		IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
		LinkAddressesFunc:    func(ctx sdk.Context, s nexus.CrossChainAddress, r nexus.CrossChainAddress) error { return nil },
		GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			c, ok := chains[chain]
			return c, ok
		},
		IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return true },
	}
	multisigKeeper := &mock.MultisigKeeperMock{
		GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
			return multisigTestUtils.KeyID(), true
		},
	}
	server := keeper.NewMsgServerImpl(k, n, &mock.VoterMock{}, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, multisigKeeper)
	_, err = server.Link(sdk.WrapSDKContext(ctx), &types.LinkRequest{Sender: rand.AccAddr(), Chain: evmChain, RecipientAddr: recipient.Address, RecipientChain: recipient.Chain.Name, Asset: axelarnet.NativeAsset})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.IsAssetRegisteredCalls()))
	assert.Equal(t, 2, len(n.GetChainCalls()))
	assert.Equal(t, 1, len(n.LinkAddressesCalls()))
	assert.Equal(t, sender, n.LinkAddressesCalls()[0].Sender)
	assert.Equal(t, recipient, n.LinkAddressesCalls()[0].Recipient)

	expected := types.BurnerInfo{BurnerAddress: types.Address(burnAddr), TokenAddress: token.GetAddress(), DestinationChain: recipient.Chain.Name, Symbol: msg.TokenDetails.Symbol, Asset: axelarnet.NativeAsset, Salt: types.Hash(salt)}
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
		ctx            sdk.Context
		basek          *mock.BaseKeeperMock
		chaink         *mock.ChainKeeperMock
		v              *mock.VoterMock
		n              *mock.NexusMock
		multisigKeeper *mock.MultisigKeeperMock
		msg            *types.ConfirmTokenRequest
		token          types.ERC20Token
		server         types.MsgServiceServer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				if chain.Equals(evmChain) {
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
			GetGatewayAddressFunc: func(sdk.Context) (types.Address, bool) {
				return types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))), true
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
			InitializePollFunc: func(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error) { return 0, nil },
		}
		chains := map[nexus.ChainName]nexus.Chain{axelarnet.Axelarnet.Name: axelarnet.Axelarnet, exported.Ethereum.Name: exported.Ethereum}
		n = &mock.NexusMock{
			GetChainMaintainersFunc: func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress {
				return []sdk.ValAddress{}
			},
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}
		multisigKeeper = &mock.MultisigKeeperMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
				return multisigTestUtils.KeyID(), true
			},
		}

		token = createMockERC20Token(axelarnet.NativeAsset, createDetails(randomNormalizedStr(10), randomNormalizedStr(3)))
		msg = &types.ConfirmTokenRequest{
			Sender: rand.AccAddr(),
			Chain:  evmChain,
			TxID:   types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Asset:  types.NewAsset(axelarnet.Axelarnet.Name.String(), axelarnet.NativeAsset),
		}
		snapshotKeeper := &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return rand.ValAddr()
			},
			CreateSnapshotFunc: func(sdk.Context, []sdk.ValAddress, func(snapshot.ValidatorI) bool, func(consensusPower sdk.Uint) sdk.Uint, utils.Threshold) (snapshot.Snapshot, error) {
				return snapshot.Snapshot{}, nil
			},
		}
		stakingKeeper := &mock.StakingKeeperMock{
			PowerReductionFunc: func(ctx sdk.Context) sdk.Int { return sdk.OneInt() },
		}
		server = keeper.NewMsgServerImpl(basek, n, v, snapshotKeeper, stakingKeeper, &mock.SlashingKeeperMock{}, multisigKeeper)
	}

	repeats := 20
	t.Run("happy path confirm", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == proto.MessageName(&types.ConfirmTokenStarted{}) }), 1)
	}).Repeat(repeats))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		msg.Chain = nexus.ChainName(rand.StrBetween(5, 20))

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
		v.InitializePollFunc = func(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error) {
			return 0, fmt.Errorf("poll setup failed")
		}

		_, err := server.ConfirmToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestAddChain(t *testing.T) {
	var (
		ctx    sdk.Context
		basek  *mock.BaseKeeperMock
		chaink *mock.ChainKeeperMock
		n      *mock.NexusMock
		msg    *types.AddChainRequest
		server types.MsgServiceServer
		name   nexus.ChainName
		params types.Params
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		chains := map[nexus.ChainName]nexus.Chain{
			exported.Ethereum.Name:   exported.Ethereum,
			axelarnet.Axelarnet.Name: axelarnet.Axelarnet,
		}
		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(n nexus.ChainName) types.ChainKeeper {
				if n == name {
					return chaink
				}
				return nil
			},
		}
		chaink = &mock.ChainKeeperMock{
			SetParamsFunc: func(sdk.Context, types.Params) {},
		}

		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
			GetChainByNativeAssetFunc: func(ctx sdk.Context, denom string) (nexus.Chain, bool) { return nexus.Chain{}, false },
			SetChainFunc:              func(sdk.Context, nexus.Chain) {},
		}

		name = nexus.ChainName(rand.StrBetween(5, 20))
		params = types.DefaultParams()[0]
		params.Chain = name
		msg = &types.AddChainRequest{
			Sender: rand.AccAddr(),
			Name:   name,
			Params: params,
		}

		server = keeper.NewMsgServerImpl(basek, n, &mock.VoterMock{}, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, &mock.MultisigKeeperMock{})
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

		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == proto.MessageName(&types.ChainAdded{}) }), 1)

	}).Repeat(repeats))

	t.Run("chain already registered", testutils.Func(func(t *testing.T) {
		setup()

		msg.Name = axelarnet.Axelarnet.Name

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestHandleMsgConfirmDeposit(t *testing.T) {
	var (
		ctx            sdk.Context
		basek          *mock.BaseKeeperMock
		chaink         *mock.ChainKeeperMock
		v              *mock.VoterMock
		multisigKeeper *mock.MultisigKeeperMock
		n              *mock.NexusMock
		msg            *types.ConfirmDepositRequest
		server         types.MsgServiceServer
	)

	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				if chain.Equals(evmChain) {
					return chaink
				}
				return nil
			},
		}

		token := types.CreateERC20Token(func(m types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
			Asset:        rand.StrBetween(5, 10),
			BurnerCode:   burnerBC,
			IsExternal:   rand.Bools(float64(0.5)).Next(),
			TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Status:       types.Confirmed,
		})

		salt := types.Hash(common.BytesToHash(rand.Bytes(common.HashLength)))
		gatewayAddr := types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength)))
		burnerAddress := types.Address(crypto.CreateAddress2(common.Address(gatewayAddr), salt, funcs.MustOk(token.GetBurnerCodeHash()).Bytes()))

		chaink = &mock.ChainKeeperMock{
			GetDepositFunc: func(sdk.Context, types.Hash, types.Address) (types.ERC20Deposit, types.DepositStatus, bool) {
				return types.ERC20Deposit{}, 0, false
			},
			GetBurnerAddressFunc: func(ctx sdk.Context, token types.ERC20Token, salt types.Hash, gatewayAddr types.Address) (types.Address, error) {
				_, ok := token.GetBurnerCodeHash()
				if !ok {
					return types.Address{}, fmt.Errorf("codehash not found")
				}
				return burnerAddress, nil
			},
			GetBurnerByteCodeFunc: func(ctx sdk.Context) ([]byte, bool) {
				return burnerBC, true
			},
			GetBurnerInfoFunc: func(ctx sdk.Context, address types.Address) *types.BurnerInfo {
				if burnerAddress != address {
					return nil
				}

				return &types.BurnerInfo{
					TokenAddress: token.GetAddress(),
					Asset:        token.GetAsset(),
					Symbol:       rand.StrBetween(5, 10),
					Salt:         salt,
				}
			},
			GetERC20TokenByAssetFunc: func(ctx sdk.Context, asset string) types.ERC20Token {
				if asset == token.GetAsset() {
					return token
				}
				return types.NilToken
			},
			GetGatewayAddressFunc: func(sdk.Context) (types.Address, bool) {
				return gatewayAddr, true
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) (int64, bool) { return rand.PosI64(), true },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) (uint64, bool) { return mathRand.Uint64(), true },
			GetVotingThresholdFunc: func(sdk.Context) (utils.Threshold, bool) {
				return utils.Threshold{Numerator: 15, Denominator: 100}, true
			},
			GetMinVoterCountFunc: func(sdk.Context) (int64, bool) { return 15, true },
			GetParamsFunc:        func(ctx sdk.Context) types.Params { return types.DefaultParams()[0] },
		}
		v = &mock.VoterMock{
			InitializePollFunc: func(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error) { return 0, nil },
		}
		multisigKeeper = &mock.MultisigKeeperMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
				return multisigTestUtils.KeyID(), true
			},
		}
		chains := map[nexus.ChainName]nexus.Chain{
			exported.Ethereum.Name:   exported.Ethereum,
			axelarnet.Axelarnet.Name: axelarnet.Axelarnet,
		}
		n = &mock.NexusMock{
			GetChainMaintainersFunc: func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress { return []sdk.ValAddress{} },
			IsChainActivatedFunc:    func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
		}

		msg = &types.ConfirmDepositRequest{
			Sender:        rand.AccAddr(),
			Chain:         evmChain,
			TxID:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			BurnerAddress: burnerAddress,
		}
		snapshotKeeper := &mock.SnapshotterMock{
			GetOperatorFunc: func(sdk.Context, sdk.AccAddress) sdk.ValAddress {
				return rand.ValAddr()
			},
			CreateSnapshotFunc: func(sdk.Context, []sdk.ValAddress, func(snapshot.ValidatorI) bool, func(consensusPower sdk.Uint) sdk.Uint, utils.Threshold) (snapshot.Snapshot, error) {
				return snapshot.Snapshot{}, nil
			},
		}
		stakingKeeper := &mock.StakingKeeperMock{
			PowerReductionFunc: func(ctx sdk.Context) sdk.Int { return sdk.OneInt() },
		}
		server = keeper.NewMsgServerImpl(basek, n, v, snapshotKeeper, stakingKeeper, &mock.SlashingKeeperMock{}, multisigKeeper)
	}

	repeats := 20
	t.Run("happy path confirm", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.NoError(t, err)
		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == proto.MessageName(&types.ConfirmDepositStarted{}) }), 1)
		assert.Equal(t, len(v.InitializePollCalls()), 1)
	}).Repeat(repeats))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		msg.Chain = nexus.ChainName(rand.StrBetween(5, 20))

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.ErrorContains(t, err, "not a registered chain")
	}).Repeat(repeats))

	t.Run("unknown gateway address", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetGatewayAddressFunc = func(sdk.Context) (types.Address, bool) { return types.Address{}, false }

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.ErrorContains(t, err, "gateway address not set for chain")
	}).Repeat(repeats))

	t.Run("invalid asset", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetERC20TokenByAssetFunc = func(ctx sdk.Context, asset string) types.ERC20Token { return types.NilToken }

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.ErrorContains(t, err, "is not confirmed on")
	}).Repeat(repeats))

	t.Run("invalid burner address", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetBurnerAddressFunc = func(ctx sdk.Context, token types.ERC20Token, salt types.Hash, gatewayAddr types.Address) (types.Address, error) {
			return types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))), nil
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.ErrorContains(t, err, "provided burner address")
		assert.ErrorContains(t, err, "doesn't match expected address")
	}).Repeat(repeats))

	t.Run("init poll failed", testutils.Func(func(t *testing.T) {
		setup()
		errMsg := "failed to initialize poll"
		v.InitializePollFunc = func(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error) {
			return 0, fmt.Errorf(errMsg)
		}
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.ErrorContains(t, err, errMsg)
	}).Repeat(repeats))
}

func TestHandleMsgCreateDeployToken(t *testing.T) {
	var (
		ctx            sdk.Context
		basek          *mock.BaseKeeperMock
		chaink         *mock.ChainKeeperMock
		v              *mock.VoterMock
		multisigKeeper *mock.MultisigKeeperMock
		n              *mock.NexusMock
		msg            *types.CreateDeployTokenRequest
		server         types.MsgServiceServer
	)
	setup := func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		msg = createMsgSignDeploy(createDetails(randomNormalizedStr(10), randomNormalizedStr(3)))

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(chain nexus.ChainName) types.ChainKeeper {
				if chain.Equals(evmChain) {
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
			GetGatewayAddressFunc: func(sdk.Context) (types.Address, bool) {
				return types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))), true
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

		chains := map[nexus.ChainName]nexus.Chain{axelarnet.Axelarnet.Name: axelarnet.Axelarnet, exported.Ethereum.Name: exported.Ethereum}
		n = &mock.NexusMock{
			IsChainActivatedFunc: func(ctx sdk.Context, chain nexus.Chain) bool { return true },
			GetChainFunc: func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				c, ok := chains[chain]
				return c, ok
			},
			IsAssetRegisteredFunc: func(sdk.Context, nexus.Chain, string) bool { return true },
			RegisterAssetFunc:     func(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset) error { return nil },
		}
		multisigKeeper = &mock.MultisigKeeperMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
				return multisigTestUtils.KeyID(), true
			},
		}

		server = keeper.NewMsgServerImpl(basek, n, v, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, multisigKeeper)
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
		msg.Chain = nexus.ChainName(rand.StrBetween(5, 20))

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when gateway is not set", testutils.Func(func(t *testing.T) {
		setup()
		chaink.GetGatewayAddressFunc = func(sdk.Context) (types.Address, bool) { return types.Address{}, false }

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when origin chain is unknown", testutils.Func(func(t *testing.T) {
		setup()
		msg.Asset.Chain = nexus.ChainName(rand.StrBetween(5, 20))

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when asset is not registered on the origin chain", testutils.Func(func(t *testing.T) {
		setup()
		n.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return false }

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("should return error when key is not set", testutils.Func(func(t *testing.T) {
		setup()
		multisigKeeper.GetCurrentKeyIDFunc = func(ctx sdk.Context, chainName nexus.ChainName) (multisig.KeyID, bool) { return "", false }

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

	ctx, msgServer, bk, n, _, _, _ := setup()
	contractCallQueue := &utilsMock.KVQueueMock{
		EnqueueFunc: func(key utils.Key, value codec.ProtoMarshaler) {},
	}
	ck = &mock.ChainKeeperMock{
		GetConfirmedEventQueueFunc: func(ctx sdk.Context) utils.KVQueue {
			return contractCallQueue
		},
	}
	bk.ForChainFunc = func(chain nexus.ChainName) types.ChainKeeper {
		return ck
	}
	bk.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }

	req := types.NewRetryFailedEventRequest(rand.AccAddr(), rand.Str(5), rand.Str(5))

	chainFound := func(found bool) func() {
		return func() {
			n.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
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

func newKeeper(ctx sdk.Context, chain nexus.ChainName, confHeight int64) types.BaseKeeper {
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
		EndBlockerLimit: 50,
	})
	k.ForChain(chain).SetGateway(ctx, types.Address(common.HexToAddress(gateway)))

	return k
}

func createMsgSignDeploy(details types.TokenDetails) *types.CreateDeployTokenRequest {
	account := rand.AccAddr()

	asset := types.NewAsset(axelarnet.Axelarnet.Name.String(), axelarnet.NativeAsset)
	return types.NewCreateDeployTokenRequest(account, exported.Ethereum.Name.String(), asset, details, types.ZeroAddress, sdk.NewUint(uint64(rand.PosI64())).String())
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
