package keeper_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	mathRand "math/rand"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/common"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	evmCrypto "github.com/ethereum/go-ethereum/crypto"
	evmParams "github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilsMock "github.com/axelarnetwork/axelar-core/utils/mock"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/evm"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	evmTestUtils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTestUtils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	multisigtypesutils "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	permission "github.com/axelarnetwork/axelar-core/x/permission/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

var (
	evmChain    = exported.Ethereum.Name
	network     = types.Ganache
	networkConf = evmParams.AllCliqueProtocolChanges
	tokenBC     = rand.Bytes(64)
	burnerBC    = common.Hex2Bytes(types.Burnable)
	gateway     = "0x37CC4B7E8f9f505CA8126Db8a9d070566ed5DAE7"
)

func setup(t log.TestingT) (sdk.Context, types.MsgServiceServer, *mock.BaseKeeperMock, *mock.NexusMock, *mock.VoterMock, *mock.SnapshotterMock, *mock.MultisigKeeperMock, *mock.PermissionMock) {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.NewTestLogger(t))

	evmBaseKeeper := &mock.BaseKeeperMock{}
	nexusKeeper := &mock.NexusMock{}
	voteKeeper := &mock.VoterMock{}
	snapshotKeeper := &mock.SnapshotterMock{}
	stakingKeeper := &mock.StakingKeeperMock{}
	slashingKeeper := &mock.SlashingKeeperMock{}
	multisigKeeper := &mock.MultisigKeeperMock{}
	permissionKeeper := &mock.PermissionMock{}

	return ctx,
		keeper.NewMsgServerImpl(evmBaseKeeper, nexusKeeper, voteKeeper, snapshotKeeper, stakingKeeper, slashingKeeper, multisigKeeper, permissionKeeper),
		evmBaseKeeper, nexusKeeper, voteKeeper, snapshotKeeper, multisigKeeper, permissionKeeper
}

func TestSetGateway(t *testing.T) {
	req := types.NewSetGatewayRequest(rand.AccAddr(), rand.Str(5), evmTestUtils.RandomAddress())

	t.Run("should fail if current key is not set", func(t *testing.T) {
		ctx, msgServer, _, nexusKeeper, _, _, multisigKeeper, _ := setup(t)

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
	})

	t.Run("should fail if gateway is already set", func(t *testing.T) {
		ctx, msgServer, baseKeeper, nexusKeeper, _, _, multisigKeeper, _ := setup(t)
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
		baseKeeper.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) { return chainKeeper, nil }
		chainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) {
			return evmTestUtils.RandomAddress(), true
		}

		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gateway already set")
	})

	t.Run("should set gateway", func(t *testing.T) {
		ctx, msgServer, baseKeeper, nexusKeeper, _, _, multisigKeeper, _ := setup(t)
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
		baseKeeper.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) { return chainKeeper, nil }
		chainKeeper.GetGatewayAddressFunc = func(ctx sdk.Context) (types.Address, bool) {
			return types.Address{}, false
		}
		chainKeeper.SetGatewayFunc = func(ctx sdk.Context, address types.Address) {}

		_, err := msgServer.SetGateway(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)
		assert.Len(t, chainKeeper.SetGatewayCalls(), 1)
		assert.Equal(t, req.Address, chainKeeper.SetGatewayCalls()[0].Address)
	})
}

func TestUpdateParams(t *testing.T) {
	ctx, msgServer, baseKeeper, _, _, _, _, _ := setup(t)

	ck := &mock.ChainKeeperMock{
		SetParamsFunc: func(ctx sdk.Context, params types.Params) {},
	}
	baseKeeper.ForChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
		require.Equal(t, chain, exported.Ethereum.Name)
		return ck, nil
	}

	p := types.DefaultParams()[0]
	p.TransferLimit = p.TransferLimit + 1
	_, err := msgServer.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: rand2.AccAddr().String(), Params: p})
	assert.NoError(t, err)
	assert.Len(t, ck.SetParamsCalls(), 1)
	assert.Equal(t, p, ck.SetParamsCalls()[0].Params)
}

func TestConfirmTransferKeyFromGovernance(t *testing.T) {
	// Mocks
	ctx, msgServer, bk, n, _, _, multisigKeeper, permissionKeeper := setup(t)
	cdc := params.MakeEncodingConfig().Codec
	confirmedEventQueue := utils.NewGeneralKVQueue(
		"q",
		utils.NewNormalizedStore(ctx.KVStore(store.NewKVStoreKey("q")), cdc),
		ctx.Logger(),
		func(value codec.ProtoMarshaler) utils.Key {
			return utils.KeyFromStr(value.String())
		},
	)
	// Capture the synthetic event produced by ConfirmTransferKey
	var capturedEvent *types.Event
	// Chain setup
	chain := nexus.Chain{Name: nexusutils.RandomChainName(), Module: types.ModuleName}
	ck := &mock.ChainKeeperMock{
		LoggerFunc:                 func(sdk.Context) log.Logger { return ctx.Logger() },
		GetParamsFunc:              func(sdk.Context) types.Params { return types.DefaultParams()[0] },
		GetConfirmedEventQueueFunc: func(sdk.Context) utils.KVQueue { return confirmedEventQueue },
		EnqueueConfirmedEventFunc: func(_ sdk.Context, id types.EventID) error {
			require.NotNil(t, capturedEvent)
			require.Equal(t, capturedEvent.GetID(), id)
			confirmedEventQueue.Enqueue(utils.LowerCaseKey(string(id)), capturedEvent)
			return nil
		},
		SetConfirmedEventFunc: func(_ sdk.Context, ev types.Event) error { capturedEvent = &ev; return nil },
		SetEventCompletedFunc: func(_ sdk.Context, id types.EventID) error { return nil },
	}
	bk.ForChainFunc = func(sdk.Context, nexus.ChainName) (types.ChainKeeper, error) { return ck, nil }
	bk.LoggerFunc = func(sdk.Context) log.Logger { return ctx.Logger() }

	n.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) { return chain, true }
	n.GetChainsFunc = func(sdk.Context) []nexus.Chain { return []nexus.Chain{chain} }
	n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return true }
	n.GetProcessingMessagesFunc = func(sdk.Context, nexus.ChainName, int64) []nexus.GeneralMessage {
		return []nexus.GeneralMessage{}
	}
	// Multisig state: pending rotation and next key
	nextKeyID := multisigTestUtils.KeyID()
	key := multisigtypesutils.Key()
	multisigKeeper.GetNextKeyIDFunc = func(sdk.Context, nexus.ChainName) (multisig.KeyID, bool) { return nextKeyID, true }
	multisigKeeper.GetKeyFunc = func(sdk.Context, multisig.KeyID) (multisig.Key, bool) { return &key, true }
	multisigKeeper.RotateKeyFunc = func(sdk.Context, nexus.ChainName) error { return nil }

	chainMgr := rand.AccAddr()
	permissionKeeper.GetRoleFunc = func(_ sdk.Context, addr sdk.AccAddress) permission.Role {
		if addr.Equals(chainMgr) {
			return permission.ROLE_CHAIN_MANAGEMENT
		}
		return permission.ROLE_UNRESTRICTED
	}

	// Force the transfer from governance; this creates and stores a synthetic event
	_, err := msgServer.ConfirmTransferKey(ctx, &types.ConfirmTransferKeyRequest{
		Chain:  chain.Name,
		Sender: chainMgr.String(),
	})
	assert.NoError(t, err)
	assert.NotNil(t, capturedEvent, "synthetic event must be captured")

	// EndBlocker should rotate the key
	_, err = evm.EndBlocker(ctx, bk, n, multisigKeeper)
	assert.NoError(t, err)
	assert.Len(t, multisigKeeper.RotateKeyCalls(), 1, "forced confirmation should rotate the key exactly once")
}

func TestSignCommands(t *testing.T) {
	setup := func() (sdk.Context, types.MsgServiceServer, *mock.BaseKeeperMock, *mock.MultisigKeeperMock) {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.NewTestLogger(t))

		evmBaseKeeper := &mock.BaseKeeperMock{}
		nexusKeeper := &mock.NexusMock{}
		voteKeeper := &mock.VoterMock{}
		snapshotKeeper := &mock.SnapshotterMock{}
		stakingKeeper := &mock.StakingKeeperMock{}
		slashingKeeper := &mock.SlashingKeeperMock{}
		multisigKeeper := &mock.MultisigKeeperMock{}
		permissionKeeper := &mock.PermissionMock{}

		nexusKeeper.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, true }
		nexusKeeper.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }

		msgServer := keeper.NewMsgServerImpl(evmBaseKeeper, nexusKeeper, voteKeeper, snapshotKeeper, stakingKeeper, slashingKeeper, multisigKeeper, permissionKeeper)

		return ctx, msgServer, evmBaseKeeper, multisigKeeper
	}

	t.Run("should create a new command batch to sign if the latest is not being signed or aborted", func(t *testing.T) {
		ctx, msgServer, evmBaseKeeper, multisigKeeper := setup()

		expectedCommandIDs := make([]types.CommandID, rand.I64Between(1, 100))
		for i := range expectedCommandIDs {
			expectedCommandIDs[i] = types.NewCommandID(rand.Bytes(common.HashLength), math.NewInt(0))
		}
		expected := types.CommandBatchMetadata{
			ID:         rand.Bytes(common.HashLength),
			CommandIDs: expectedCommandIDs,
			Status:     types.BatchSigning,
			KeyID:      multisigTestUtils.KeyID(),
		}

		chainKeeper := &mock.ChainKeeperMock{}
		evmBaseKeeper.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
		evmBaseKeeper.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) { return chainKeeper, nil }
		chainKeeper.GetChainIDFunc = func(ctx sdk.Context) (math.Int, bool) { return math.NewInt(0), true }
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
	})

	t.Run("should get the latest if it is aborted", func(t *testing.T) {
		ctx, msgServer, evmBaseKeeper, signerKeeper := setup()

		expectedCommandIDs := make([]types.CommandID, rand.I64Between(1, 100))
		for i := range expectedCommandIDs {
			expectedCommandIDs[i] = types.NewCommandID(rand.Bytes(common.HashLength), math.NewInt(0))
		}
		commandBatch := types.CommandBatchMetadata{
			ID:         rand.Bytes(common.HashLength),
			CommandIDs: expectedCommandIDs,
			Status:     types.BatchAborted,
			KeyID:      multisigTestUtils.KeyID(),
		}

		chainKeeper := &mock.ChainKeeperMock{}
		evmBaseKeeper.LoggerFunc = func(ctx sdk.Context) log.Logger { return ctx.Logger() }
		evmBaseKeeper.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) { return chainKeeper, nil }
		chainKeeper.GetChainIDFunc = func(ctx sdk.Context) (math.Int, bool) { return math.NewInt(0), true }
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
	})
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
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.NewTestLogger(t))

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				if chain.Equals(evmChain) {
					return chaink, nil
				}
				return nil, errors.New("unknown chain")
			},
		}
		chaink = &mock.ChainKeeperMock{
			GetVotingThresholdFunc: func(sdk.Context) utils.Threshold {
				return utils.Threshold{Numerator: 15, Denominator: 100}
			},
			GetMinVoterCountFunc: func(sdk.Context) int64 { return 15 },
			GetGatewayAddressFunc: func(sdk.Context) (types.Address, bool) {
				return types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))), true
			},
			GetRevoteLockingPeriodFunc:        func(sdk.Context) int64 { return rand.PosI64() },
			GetRequiredConfirmationHeightFunc: func(sdk.Context) uint64 { return mathRand.Uint64() },
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
			Sender: rand.AccAddr().String(),
			Chain:  evmChain,
			TxID:   types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
			Asset:  types.NewAsset(axelarnet.Axelarnet.Name.String(), axelarnet.NativeAsset),
		}
		snapshotKeeper := &mock.SnapshotterMock{
			CreateSnapshotFunc: func(sdk.Context, []sdk.ValAddress, func(snapshot.ValidatorI) bool, func(consensusPower math.Uint) math.Uint, utils.Threshold) (snapshot.Snapshot, error) {
				return snapshot.Snapshot{}, nil
			},
		}
		stakingKeeper := &mock.StakingKeeperMock{
			PowerReductionFunc: func(context.Context) math.Int { return math.OneInt() },
		}
		server = keeper.NewMsgServerImpl(basek, n, v, snapshotKeeper, stakingKeeper, &mock.SlashingKeeperMock{}, multisigKeeper, &mock.PermissionMock{})
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
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.NewTestLogger(t))

		chains := map[nexus.ChainName]nexus.Chain{
			exported.Ethereum.Name:   exported.Ethereum,
			axelarnet.Axelarnet.Name: axelarnet.Axelarnet,
		}
		basek = &mock.BaseKeeperMock{
			CreateChainFunc: func(_ sdk.Context, params types.Params) error { return nil },
			ForChainFunc: func(_ sdk.Context, n nexus.ChainName) (types.ChainKeeper, error) {
				if n == name {
					return chaink, nil
				}
				return nil, errors.New("unknown chain")
			},
		}
		chaink = &mock.ChainKeeperMock{}

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
			Sender: rand.AccAddr().String(),
			Name:   name,
			Params: params,
		}

		server = keeper.NewMsgServerImpl(basek, n, &mock.VoterMock{}, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, &mock.MultisigKeeperMock{}, &mock.PermissionMock{})
	}

	repeats := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(n.SetChainCalls()))
		assert.Equal(t, name, n.SetChainCalls()[0].Chain.Name)

		_, err = basek.ForChain(ctx, name)
		assert.NoError(t, err)

		assert.Len(t, testutils.Events(ctx.EventManager().ABCIEvents()).Filter(func(event abci.Event) bool { return event.Type == proto.MessageName(&types.ChainAdded{}) }), 1)

	}).Repeat(repeats))

	t.Run("chain already registered", testutils.Func(func(t *testing.T) {
		setup()

		msg.Name = axelarnet.Axelarnet.Name

		_, err := server.AddChain(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
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
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.NewTestLogger(t))
		msg = createMsgSignDeploy(createDetails(randomNormalizedStr(10), randomNormalizedStr(3)))

		basek = &mock.BaseKeeperMock{
			ForChainFunc: func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) {
				if chain == evmChain {
					return chaink, nil
				}
				return nil, errors.New("unknown chain")
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
			GetChainIDByNetworkFunc: func(ctx sdk.Context, network string) (math.Int, bool) {
				return math.NewInt(rand.I64Between(1, 1000)), true
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
			RegisterAssetFunc: func(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset, limit math.Uint, window time.Duration) error {
				return nil
			},
		}
		multisigKeeper = &mock.MultisigKeeperMock{
			GetCurrentKeyIDFunc: func(ctx sdk.Context, chain nexus.ChainName) (multisig.KeyID, bool) {
				return multisigTestUtils.KeyID(), true
			},
		}

		server = keeper.NewMsgServerImpl(basek, n, v, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.SlashingKeeperMock{}, multisigKeeper, &mock.PermissionMock{})
	}

	repeats := 20
	t.Run("should create deploy token when gateway address is set, chains are registered and asset is registered on the origin chain ", testutils.Func(func(t *testing.T) {
		setup()

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chaink.EnqueueCommandCalls()))
	}).Repeat(repeats))

	t.Run("should create deploy token with infinite rate limit", testutils.Func(func(t *testing.T) {
		setup()
		msg.DailyMintLimit = "0"

		_, err := server.CreateDeployToken(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chaink.EnqueueCommandCalls()))
	}))

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

	ctx, msgServer, bk, n, _, _, _, _ := setup(t)
	contractCallQueue := &utilsMock.KVQueueMock{
		EnqueueFunc: func(key utils.Key, value codec.ProtoMarshaler) {},
	}
	ck = &mock.ChainKeeperMock{
		GetConfirmedEventQueueFunc: func(ctx sdk.Context) utils.KVQueue {
			return contractCallQueue
		},
	}
	bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) { return ck, nil }
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
				return types.Event{
					Status: eventStatus,
					Event: &types.Event_ContractCall{
						ContractCall: &types.EventContractCall{},
					}}, true
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

func TestHandleMsgConfirmGatewayTxs(t *testing.T) {
	validators := slices.Expand(func(int) snapshot.Participant { return snapshot.NewParticipant(rand2.ValAddr(), math.OneUint()) }, 10)
	txIDs := slices.Expand2(evmTestUtils.RandomHash, int(rand.I64Between(5, 50)))
	req := types.NewConfirmGatewayTxsRequest(rand.AccAddr(), nexus.ChainName(rand.Str(5)), txIDs)

	var (
		ctx         sdk.Context
		bk          *mock.BaseKeeperMock
		ck          *mock.ChainKeeperMock
		s           *mock.SlashingKeeperMock
		n           *mock.NexusMock
		snapshotter *mock.SnapshotterMock
		v           *mock.VoterMock
		msgServer   types.MsgServiceServer
		pollID      vote.PollID
	)

	givenMsgServer := Given("an EVM msg server", func() {
		ctx = rand2.Context(fake.NewMultiStore(), t)

		bk = &mock.BaseKeeperMock{
			LoggerFunc:   func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			ForChainFunc: func(sdk.Context, nexus.ChainName) (types.ChainKeeper, error) { return nil, fmt.Errorf("unknown chain") },
		}
		snapshotter = &mock.SnapshotterMock{
			CreateSnapshotFunc: func(sdk.Context, []sdk.ValAddress, func(snapshot.ValidatorI) bool, func(consensusPower math.Uint) math.Uint, utils.Threshold) (snapshot.Snapshot, error) {
				return snapshot.NewSnapshot(ctx.BlockTime(), ctx.BlockHeight(), validators, math.NewUint(10)), nil
			},
		}
		ck = &mock.ChainKeeperMock{
			GetRequiredConfirmationHeightFunc: func(sdk.Context) uint64 { return 10 },
			GetParamsFunc:                     func(sdk.Context) types.Params { return types.DefaultParams()[0] },
		}
		n = &mock.NexusMock{
			GetChainMaintainersFunc: func(sdk.Context, nexus.Chain) []sdk.ValAddress {
				return slices.Expand2(rand2.ValAddr, 10)
			},
		}
		s = &mock.SlashingKeeperMock{}
		v = &mock.VoterMock{}
		pollID = vote.PollID(0)

		msgServer = keeper.NewMsgServerImpl(bk, n, v, snapshotter, &mock.StakingKeeperMock{}, s, &mock.MultisigKeeperMock{}, &mock.PermissionMock{})
	})

	whenChainIsValid := When("chain is set and activated", func() {
		n.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, true }
		n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return true }
		bk.ForChainFunc = func(_ sdk.Context, chain nexus.ChainName) (types.ChainKeeper, error) { return ck, nil }
		ck.GetGatewayAddressFunc = func(sdk.Context) (types.Address, bool) { return evmTestUtils.RandomAddress(), true }
	})

	whenSnapshotIsCreated := When("snapshot is created", func() {
		snapshotter.GetProxyFunc = func(sdk.Context, sdk.ValAddress) (sdk.AccAddress, bool) {
			return rand2.AccAddr(), true
		}
		s.IsTombstonedFunc = func(context.Context, sdk.ConsAddress) bool { return false }
	})

	whenPollsAreInitialized := When("polls are initialized", func() {
		v.InitializePollFunc = func(sdk.Context, vote.PollBuilder) (vote.PollID, error) {
			pollID += 1
			return pollID, nil
		}
	})

	t.Run("confirm gateway txs", func(t *testing.T) {
		givenMsgServer.Branch(
			whenChainIsValid.
				When("failed to create snapshot", func() {
					snapshotter.CreateSnapshotFunc = func(sdk.Context, []sdk.ValAddress, func(snapshot.ValidatorI) bool, func(consensusPower math.Uint) math.Uint, utils.Threshold) (snapshot.Snapshot, error) {
						return snapshot.Snapshot{}, fmt.Errorf("failed to create snapshot")
					}
				}).
				Then("should return error", func(t *testing.T) {
					_, err := msgServer.ConfirmGatewayTxs(sdk.WrapSDKContext(ctx), req)
					assert.ErrorContains(t, err, "failed to create snapshot")
				}),
			whenChainIsValid.
				When2(whenSnapshotIsCreated).
				When("failed to initialize polls", func() {
					v.InitializePollFunc = func(sdk.Context, vote.PollBuilder) (vote.PollID, error) {
						return 0, fmt.Errorf("failed to initialize polls")
					}
				}).
				Then("should return error", func(t *testing.T) {
					_, err := msgServer.ConfirmGatewayTxs(sdk.WrapSDKContext(ctx), req)
					assert.ErrorContains(t, err, "failed to initialize polls")
				}),
			whenChainIsValid.
				When2(whenSnapshotIsCreated).
				When2(whenPollsAreInitialized).
				Then("should emit ConfirmGatewayTxsEvent", func(t *testing.T) {
					_, err := msgServer.ConfirmGatewayTxs(sdk.WrapSDKContext(ctx), req)
					assert.Equal(t, 1, len(ctx.EventManager().Events()))
					assert.NoError(t, err)
				}),
		).Run(t)
	})
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
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("subspace"), store.NewKVStoreKey("tsubspace"))
	k := keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("testKey"), paramsK)
	k.InitChains(ctx)
	funcs.MustNoErr(k.CreateChain(ctx, types.Params{
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
			Id:   math.NewIntFromUint64(uint64(rand.I64Between(1, 10))),
		}},
		EndBlockerLimit: 50,
		TransferLimit:   50,
	}))
	funcs.Must(k.ForChain(ctx, chain)).SetGateway(ctx, types.Address(common.HexToAddress(gateway)))

	return k
}

func createMsgSignDeploy(details types.TokenDetails) *types.CreateDeployTokenRequest {
	account := rand.AccAddr()

	asset := types.NewAsset(axelarnet.Axelarnet.Name.String(), axelarnet.NativeAsset)
	return types.NewCreateDeployTokenRequest(account, exported.Ethereum.Name.String(), asset, details, types.ZeroAddress, math.NewUint(uint64(rand.PosI64())).String())
}

func createDetails(name, symbol string) types.TokenDetails {
	decimals := rand.Bytes(1)[0]
	capacity := math.NewIntFromUint64(uint64(rand.PosI64()))

	return types.NewTokenDetails(name, symbol, decimals, capacity)
}

func createMockERC20Token(asset string, details types.TokenDetails) types.ERC20Token {
	meta := types.ERC20TokenMetadata{
		Asset:        asset,
		Details:      details,
		Status:       types.Initialized,
		TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
		ChainID:      math.NewIntFromUint64(uint64(rand.I64Between(1, 10))),
	}
	return types.CreateERC20Token(
		func(meta types.ERC20TokenMetadata) {},
		meta,
	)
}

func randomNormalizedStr(size int) string {
	return strings.ReplaceAll(utils.NormalizeString(rand.Str(size)), utils.DefaultDelimiter, "-")
}
