package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisigTestUtils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	types2 "github.com/axelarnetwork/axelar-core/x/multisig/types"
	testutils2 "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestCommands(t *testing.T) {
	var (
		ctx   sdk.Context
		k     *evmKeeper.BaseKeeper
		chain nexus.ChainName
	)

	setup := func() {
		encCfg := params.MakeEncodingConfig()
		encCfg.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &types2.MultiSig{})
		paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("params"), store.NewKVStoreKey("tparams"))
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
		k = evmKeeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("evm"), paramsK)
		k.InitChains(ctx)
		chain = "Ethereum"
	}

	repeats := 20

	t.Run("enqueue and then batch commands", testutils.Func(func(t *testing.T) {
		setup()
		assert.NoError(t, k.CreateChain(ctx, types.DefaultParams()[0]))
		chainKeeper := funcs.Must(k.ForChain(ctx, chain))
		chainID, ok := chainKeeper.GetChainID(ctx)
		assert.True(t, ok)

		numCmds := int(rand.I64Between(10, 50))
		var commands []types.Command

		for i := 0; i < numCmds; i++ {
			tokenDetails := createDetails(rand.NormalizedStr(10), rand.NormalizedStr(10))
			cmd := types.NewDeployTokenCommand(chainID, multisigTestUtils.KeyID(), rand.Str(5), tokenDetails, types.ZeroAddress, math.NewUint(uint64(rand.PosI64())))

			err := chainKeeper.EnqueueCommand(ctx, cmd)
			assert.NoError(t, err)

			commands = append(commands, cmd)
		}

		for _, cmd := range commands {
			fetchedCmd, ok := chainKeeper.GetCommand(ctx, cmd.ID)
			assert.True(t, ok)
			assert.Equal(t, cmd, fetchedCmd)
		}
		assert.ElementsMatch(t, commands, chainKeeper.GetPendingCommands(ctx))

		lastLength := len(chainKeeper.GetPendingCommands(ctx))
		for {
			_, err := chainKeeper.CreateNewBatchToSign(ctx)
			assert.NoError(t, err)
			remainingCmds := chainKeeper.GetPendingCommands(ctx)
			assert.Less(t, len(remainingCmds), lastLength)
			lastLength = len(remainingCmds)
			batch := chainKeeper.GetLatestCommandBatch(ctx)
			sig := testutils2.MultiSig()
			assert.NoError(t, batch.SetSigned(&sig))
			if lastLength == 0 {
				break
			}
		}
	}).Repeat(repeats))
}

func TestGetTokenAddress(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("subspace"), store.NewKVStoreKey("tsubspace"))
	k := evmKeeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("testKey"), paramsK)
	k.InitChains(ctx)

	chain := nexus.ChainName("Ethereum")
	asset := "axelar"
	tokenName := "axelar token"
	tokenSymbol := "at"
	decimals := uint8(18)
	capacity := math.NewIntFromUint64(uint64(10000))

	axelarGateway := types.Address(common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA"))
	expected := "0x7779c3e9a8b1856b4E3Ab40da37200dbd007d594"

	funcs.MustNoErr(k.CreateChain(ctx, types.DefaultParams()[0]))
	keeper := funcs.Must(k.ForChain(ctx, chain))
	keeper.SetGateway(ctx, axelarGateway)
	tokenDetails := types.NewTokenDetails(tokenName, tokenSymbol, decimals, capacity)
	token, err := keeper.CreateERC20Token(ctx, asset, tokenDetails, types.ZeroAddress)
	assert.NoError(t, err)
	assert.Equal(t, expected, token.GetAddress().Hex())
}

func TestBaseKeeper(t *testing.T) {
	var (
		evmStoreKey       *store.KVStoreKey
		keeper            *evmKeeper.BaseKeeper
		ctx               sdk.Context
		expectedChainName nexus.ChainName
		paramstoreKey     = store.NewKVStoreKey(paramstypes.StoreKey)
		paramTStoreKey    = store.NewKVStoreKey(paramstypes.TStoreKey)
	)
	givenBaseKeeper := Given("a base keeper", func() {
		encodingConfig := app.MakeEncodingConfig()
		pKeeper := paramsKeeper.NewKeeper(encodingConfig.Codec, encodingConfig.Amino, paramstoreKey, paramTStoreKey)
		evmStoreKey = store.NewKVStoreKey(types.StoreKey)
		keeper = evmKeeper.NewKeeper(encodingConfig.Codec, evmStoreKey, pKeeper)
	})

	givenCtx := Given("a context", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
	})

	givenNoChainsExist := Given("no chains exist", func() {})

	givenChainsExistInStore := Given("a chain exists in the store", func() {
		// use local keeper to simulate shut down after chain is created
		encodingConfig := app.MakeEncodingConfig()
		pKeeper := paramsKeeper.NewKeeper(encodingConfig.Codec, encodingConfig.Amino, paramstoreKey, paramTStoreKey)
		localKeeper := evmKeeper.NewKeeper(encodingConfig.Codec, evmStoreKey, pKeeper)
		localKeeper.InitChains(ctx)

		ps := types.DefaultParams()[0]
		expectedChainName = nexustestutils.RandomChainName()
		ps.Chain = expectedChainName
		funcs.MustNoErr(localKeeper.CreateChain(ctx, ps))
	})

	whenInitChains := When("initializing chains", func() {
		keeper.InitChains(ctx)
	})

	whenNotInitialized := When("the keeper is not initialized", func() {})

	thenChainKeeperPanics := Then("the chain keeper functions panic", func(t *testing.T) {
		assert.Panics(t, func() { funcs.MustNoErr(keeper.CreateChain(ctx, types.DefaultParams()[0])) })
		assert.Panics(t, func() { _ = funcs.Must(keeper.ForChain(ctx, types.DefaultParams()[0].Chain)) })
	})

	thenNewChainCanBeCreated := Then("a new chain can be created", func(t *testing.T) {
		ps := types.DefaultParams()[0]
		expectedChainName = nexustestutils.RandomChainName()
		ps.Chain = expectedChainName
		assert.NoError(t, keeper.CreateChain(ctx, ps))
	})

	thenCreateExistingChainFails := Then("creating a new chain colliding with an existing chain fails", func(t *testing.T) {
		ps := types.DefaultParams()[0]
		ps.Chain = expectedChainName
		assert.Error(t, keeper.CreateChain(ctx, ps))
	})

	thenGetChainKeeperForExistingChain := Then("the chain keeper for an existing chain is accessible", func(t *testing.T) {
		ck, err := keeper.ForChain(ctx, expectedChainName)
		assert.NoError(t, err)
		assert.NotPanics(t, func() {
			_ = ck.GetParams(ctx)
		})
	})

	thenGetChainKeeperForUnkownChainFails := Then("getting the chain keeper for an unknown chain fails", func(t *testing.T) {
		_, err := keeper.ForChain(ctx, nexustestutils.RandomChainName())
		assert.Error(t, err)
	})

	thenSecondInitChainsPanics := Then("a second chain initialization panics", func(t *testing.T) {
		assert.Panics(t, func() {
			keeper.InitChains(ctx)
		})
	})

	thenNewChainWithWrongParamsFails := Then("creating a new chain colliding with invalid parameters fails", func(t *testing.T) {
		ps := types.DefaultParams()[0]
		ps.Network = rand.Str(5)
		assert.Error(t, ps.Validate())
		assert.Error(t, keeper.CreateChain(ctx, ps))
	})

	givenBaseKeeper.
		Given2(givenCtx).
		When2(whenNotInitialized).
		Then2(thenChainKeeperPanics).Run(t)

	givenBaseKeeper.
		Given2(givenCtx).
		Given2(givenNoChainsExist).
		When2(whenInitChains).
		Then2(thenNewChainCanBeCreated).
		Then2(thenGetChainKeeperForExistingChain).Run(t)

	givenBaseKeeper.
		Given2(givenCtx).
		Given2(givenChainsExistInStore).
		When2(whenInitChains).
		Then2(thenCreateExistingChainFails).
		Then2(thenGetChainKeeperForExistingChain).Run(t)

	givenBaseKeeper.
		Given2(givenCtx).
		When2(whenInitChains).
		Then2(thenSecondInitChainsPanics).Run(t)

	givenBaseKeeper.
		Given2(givenCtx).
		Given2(givenChainsExistInStore).
		When2(whenInitChains).
		Then2(thenGetChainKeeperForUnkownChainFails).Run(t)

	givenBaseKeeper.
		Given2(givenCtx).
		Given2(givenNoChainsExist).
		When2(whenInitChains).
		Then2(thenNewChainWithWrongParamsFails).Run(t)
}
