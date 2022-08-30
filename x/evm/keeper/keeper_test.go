package keeper_test

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/evm/exported"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisigTestUtils "github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func TestCommands(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper types.BaseKeeper
		chain  nexus.ChainName
	)

	setup := func() {
		encCfg := params.MakeEncodingConfig()
		paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = evmKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("evm"), paramsK)
		chain = "Ethereum"
	}

	repeats := 20

	t.Run("enqueue and then batch commands", testutils.Func(func(t *testing.T) {
		setup()
		chainKeeper := keeper.ForChain(chain)
		chainKeeper.SetParams(ctx, types.DefaultParams()[0])
		chainID, ok := chainKeeper.GetChainID(ctx)
		assert.True(t, ok)

		numCmds := int(rand.I64Between(10, 50))
		var commands []types.Command

		for i := 0; i < numCmds; i++ {
			tokenDetails := createDetails(rand.NormalizedStr(10), rand.NormalizedStr(10))
			cmd, err := types.CreateDeployTokenCommand(chainID, multisigTestUtils.KeyID(), rand.Str(5), tokenDetails, types.ZeroAddress, sdk.NewUint(uint64(rand.PosI64())))
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
	}).Repeat(repeats))
}

func TestSetBurnerInfoGetBurnerInfo(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper types.BaseKeeper
		chain  nexus.ChainName
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
		actual := keeper.ForChain(chain).GetBurnerInfo(ctx, burnerInfo.BurnerAddress)

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

func TestGetTokenAddress(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := evmKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("testKey"), paramsK)

	chain := nexus.ChainName("Ethereum")
	asset := "axelar"
	tokenName := "axelar token"
	tokenSymbol := "at"
	decimals := uint8(18)
	capacity := sdk.NewIntFromUint64(uint64(10000))

	axelarGateway := types.Address(common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA"))
	expected := "0x7779c3e9a8b1856b4E3Ab40da37200dbd007d594"

	keeper := k.ForChain(chain)
	keeper.SetParams(ctx, types.DefaultParams()[0])
	keeper.SetGateway(ctx, axelarGateway)
	tokenDetails := types.NewTokenDetails(tokenName, tokenSymbol, decimals, capacity)
	token, err := keeper.CreateERC20Token(ctx, asset, tokenDetails, types.ZeroAddress)
	assert.NoError(t, err)
	assert.Equal(t, expected, token.GetAddress().Hex())
}

func TestGetBurnerAddress(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
	ctx = ctx.WithHeaderHash(common.Hex2Bytes("CA36CA3751A5B6E8B8ED4072BFA5E6E5BAC8B6E06E02DE029E1BD86AB141F2F1"))
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(1000000))
	ctx.GasMeter().ConsumeGas(1000, "test")
	paramsK := paramsKeeper.NewKeeper(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("subspace"), sdk.NewKVStoreKey("tsubspace"))
	k := evmKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("testKey"), paramsK)

	bzBurnable, err := hex.DecodeString(types.Burnable)
	if err != nil {
		panic(err)
	}

	t.Run("should work for internal erc20 tokens", testutils.Func(func(t *testing.T) {
		axelarGateway := types.Address(common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA"))
		recipient := "1KDeqnsTRzFeXRaENA6XLN1EwdTujchr4L"
		tokenAddr := types.Address(common.HexToAddress("0xE7481ECB61F9C84b91C03414F3D5d48E5436045D"))
		expectedBurnerAddr := "0x294C0419D756F7C31A521659f9b3EA7a7575d4b0"
		expectedSalt := common.Hex2Bytes("b365d534cb5d28d511a8baf1125240c97b09cb46710645b30ed64f302c4ae7ff")

		chainKeeper := k.ForChain(exported.Ethereum.Name)
		chainKeeper.SetParams(ctx, types.DefaultParams()[0])
		token := types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
			TokenAddress: tokenAddr,
			IsExternal:   false,
			BurnerCode:   bzBurnable,
		})
		actualSalt := chainKeeper.GenerateSalt(ctx, recipient)
		actualburnerAddr, err := chainKeeper.GetBurnerAddress(ctx, token, actualSalt, axelarGateway)

		assert.NoError(t, err)
		assert.Equal(t, expectedBurnerAddr, actualburnerAddr.Hex())
		assert.Equal(t, common.Bytes2Hex(expectedSalt), common.Bytes2Hex(actualSalt[:]))
	}))

	t.Run("should work for external erc20 tokens", testutils.Func(func(t *testing.T) {
		axelarGateway := types.Address(common.HexToAddress("0xA193E42526F1FEA8C99AF609dcEabf30C1c29fAA"))
		recipient := "axelar1aguuy756cpaqnfd5t5qn68u7ck7w2sp64023hk"
		tokenAddr := types.Address(common.HexToAddress("0xFDFEF9D10d929cB3905C71400ce6be1990EA0F34"))
		expectedBurnerAddr := "0x3EF0e1bdF7A9c239016ce3904eAc4f458C1503D7"
		expectedSalt := common.Hex2Bytes("2321c4ff5401853a7a9960fd93a0281cde689966a62d049bdc5c5b16733954f1")

		chainKeeper := k.ForChain(exported.Ethereum.Name)
		chainKeeper.SetParams(ctx, types.DefaultParams()[0])
		token := types.CreateERC20Token(func(meta types.ERC20TokenMetadata) {}, types.ERC20TokenMetadata{
			TokenAddress: tokenAddr,
			IsExternal:   true,
			BurnerCode:   nil,
		})
		actualSalt := chainKeeper.GenerateSalt(ctx, recipient)
		actualburnerAddr, err := chainKeeper.GetBurnerAddress(ctx, token, actualSalt, axelarGateway)

		assert.NoError(t, err)
		assert.Equal(t, expectedBurnerAddr, actualburnerAddr.Hex())
		assert.Equal(t, common.Bytes2Hex(expectedSalt), common.Bytes2Hex(actualSalt[:]))
	}))
}
