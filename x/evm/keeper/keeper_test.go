package keeper_test

import (
	"testing"

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

func TestSetBurnerInfoGetBurnerInfo(t *testing.T) {
	var (
		ctx    sdk.Context
		keeper types.BaseKeeper
		chain  string
	)

	setup := func() {
		encCfg := params.MakeEncodingConfig()
		paramsK := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("params"), sdk.NewKVStoreKey("tparams"))
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = evmKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("evm"), paramsK)
		chain = "Ethereum"
	}

	t.Run("should set and get the burner info", testutils.Func(func(t *testing.T) {
		setup()

		burnerInfo := types.BurnerInfo{
			TokenAddress: types.Address(common.BytesToAddress(rand.Bytes(common.AddressLength))),
			Symbol:       rand.StrBetween(2, 5),
			Salt:         types.Hash(common.BytesToHash(rand.Bytes(common.HashLength))),
		}
		burnerAddress := common.BytesToAddress(rand.Bytes(common.AddressLength))

		keeper.ForChain(chain).SetBurnerInfo(ctx, burnerAddress, &burnerInfo)
		actual := keeper.ForChain(chain).GetBurnerInfo(ctx, burnerAddress)

		assert.NotNil(t, actual)
		assert.Equal(t, *actual, burnerInfo)
	}).Repeat(20))

}

func TestKeeper_GetParams(t *testing.T) {
	var (
		keeperWithSubspace    types.BaseKeeper
		keeperWithoutSubspace types.BaseKeeper
		ctx                   sdk.Context
	)
	setup := func() {
		encCfg := params.MakeEncodingConfig()

		// store keys need to be the same instance for all keepers, otherwise ctx will create a new underlying store,
		// even though the key string is the same
		paramStoreKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
		paramTStoreKey := sdk.NewKVStoreKey(paramstypes.TStoreKey)
		storeKey := sdk.NewKVStoreKey(types.StoreKey)

		paramsK1 := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, paramStoreKey, paramTStoreKey)
		paramsK2 := paramsKeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, paramStoreKey, paramTStoreKey)
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		keeperWithSubspace = evmKeeper.NewKeeper(encCfg.Marshaler, storeKey, paramsK1)
		keeperWithoutSubspace = evmKeeper.NewKeeper(encCfg.Marshaler, storeKey, paramsK2)

		// load params into a subspace
		keeperWithSubspace.SetParams(ctx, types.DefaultParams()...)
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
