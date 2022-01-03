package keeper_test

import (
	"testing"

	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

func TestCommands(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper types.BaseKeeper
		chain  string
	)

	setup := func() {
		encCfg := params.MakeEncodingConfig()
		paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = evmKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)
		chain = "Ethereum"
	}

	t.Run("enqueue and then batch commands", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper := keeper.ForChain(chain)
		chainKeeper.SetParams(ctx, types.DefaultParams()[0])
		chainID, ok := chainKeeper.GetChainID(ctx)
		assert.True(t, ok)

		numCmds := int(rand.I64Between(10, 50))
		var commands []types.Command

		for i := 0; i < numCmds; i++ {
			tokenDetails := createDetails(randomNormalizedStr(10), randomNormalizedStr(3))
			cmd, err := types.CreateDeployTokenCommand(chainID, tss.KeyID(rand.HexStr(10)), tokenDetails)
			assert.NoError(t, err)

			err = chainKeeper.EnqueueCommand(ctx, cmd)
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
			batch.SetStatus(types.BatchSigned)
			if lastLength == 0 {
				break
			}
		}
	}).Repeat(20))
}

func TestSetBurnerInfoGetBurnerInfo(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper types.BaseKeeper
		chain  string
	)

	setup := func() {
		encCfg := params.MakeEncodingConfig()
		paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = evmKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)
		chain = "Ethereum"
	}

	t.Run("should set and get the burner info", testutils.Func(func(t *testing.T) {
		setup()

		burnerInfo := types.BurnerInfo{
			BurnerAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			TokenAddress:  types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Symbol:        rand.StrBetween(2, 5),
			Salt:          types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		}

		keeper.ForChain(chain).SetBurnerInfo(ctx, burnerInfo)
		actual := keeper.ForChain(chain).GetBurnerInfo(ctx, common.Address(burnerInfo.BurnerAddress))

		assert.NotNil(t, actual)
		assert.Equal(t, *actual, burnerInfo)
	}).Repeat(20))

}

func TestKeeper_GetParams(t *testing.T) {
	var (
		keeperWithSubspace    types.ChainKeeper
		keeperWithoutSubspace types.ChainKeeper
		ctx                   sdk.Context
	)
	setup := func() {
		encCfg := params.MakeEncodingConfig()

		// store keys need to be the same instance for all keepers, otherwise ctx will create a new underlying store,
		// even though the key string is the same
		paramStoreKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
		paramTStoreKey := sdk.NewKVStoreKey(paramstypes.TStoreKey)
		storeKey := sdk.NewKVStoreKey(types.StoreKey)

		paramsK1 := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, paramStoreKey, paramTStoreKey)
		paramsK2 := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, paramStoreKey, paramTStoreKey)
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		keeperWithSubspace = evmKeeper.NewKeeper(encCfg.Codec, storeKey, paramsK1).ForChain(exported.Ethereum.Name)
		keeperWithoutSubspace = evmKeeper.NewKeeper(encCfg.Codec, storeKey, paramsK2).ForChain(exported.Ethereum.Name)

		// load params into a subspace
		keeperWithSubspace.SetParams(ctx, types.DefaultParams()[0])
	}

	// assert: the ctx kvstore stores all the keys of the subspace, but keeperWithoutSubspace has no Subspace created to access it
	t.Run("creating subspaces consumes no additional gas", testutils.Func(func(t *testing.T) {
		setup()

		// reset gas meter for each access
		ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
		_ = keeperWithSubspace.GetParams(ctx)
		gasWithSubspace := ctx.GasMeter().GasConsumed()

		// reset gas meter for each access
		ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
		_ = keeperWithoutSubspace.GetParams(ctx)
		gasWithoutSubspace := ctx.GasMeter().GasConsumed()

		assert.Equal(t, gasWithSubspace, gasWithoutSubspace)
	}).Repeat(20))
}
