package keeper_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnetKeeper "github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
)

func TestKeeper_GetIBCPath(t *testing.T) {
	repeats := 20

	var (
		ctx    sdk.Context
		keeper axelarnetKeeper.Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		axelarnetSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = axelarnetKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace)
	}
	t.Run("should return the registered IBC path when the given asset is registered", testutils.Func(func(t *testing.T) {
		setup()
		path := randomIBCPath()
		asset := randomDenom()
		err := keeper.RegisterIBCPath(ctx, asset, path)
		assert.NoError(t, err)
		result, ok := keeper.GetIBCPath(ctx, asset)
		assert.Equal(t, path, result)
		assert.True(t, ok)
	}).Repeat(repeats))

	t.Run("should return error when registered the same asset twice", testutils.Func(func(t *testing.T) {
		setup()
		path := randomIBCPath()
		asset := randomDenom()
		err := keeper.RegisterIBCPath(ctx, asset, path)
		assert.NoError(t, err)
		path2 := randomIBCPath()
		err2 := keeper.RegisterIBCPath(ctx, asset, path2)
		assert.Error(t, err2)
	}).Repeat(repeats))

}

func TestKeeper_RegisterCosmosChain(t *testing.T) {
	repeats := 20

	var (
		ctx    sdk.Context
		keeper axelarnetKeeper.Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		axelarnetSubspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = axelarnetKeeper.NewKeeper(encCfg.Marshaler, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace)
	}
	t.Run("should return list of registered cosmos chains", testutils.Func(func(t *testing.T) {
		setup()

		count := rand.I64Between(10, 100)
		chains := make([]string, count)

		for i := 0; i < int(count); i++ {
			chains[i] = strings.ToLower(rand.Str(10))
			keeper.RegisterAssetToCosmosChain(ctx, rand.Str(10), chains[i])
		}
		sort.Strings(chains)
		assert.Equal(t, chains, keeper.GetCosmosChains(ctx))

	}).Repeat(repeats))

	t.Run("should empty list when no chain registered", testutils.Func(func(t *testing.T) {
		setup()
		var empty []string
		assert.Equal(t, empty, keeper.GetCosmosChains(ctx))

	}).Repeat(repeats))

}

func randomIBCPath() string {
	port := rand.StrBetween(5, 10)
	identifier := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	return port + "/" + identifier
}
