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
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

func TestKeeper_GetIBCPath(t *testing.T) {
	repeats := 20

	var (
		ctx    sdk.Context
		keeper axelarnetKeeper.Keeper
	)
	setup := func() {
		encCfg := appParams.MakeEncodingConfig()
		axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = axelarnetKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace)
	}
	t.Run("should return the registered IBC path when the given asset is registered", testutils.Func(func(t *testing.T) {
		setup()
		path := randomIBCPath()
		chain := randomChain()
		chain.Assets = nil
		chain.IBCPath = ""
		keeper.SetCosmosChain(ctx, chain)
		err := keeper.RegisterIBCPath(ctx, chain.Name, path)
		assert.NoError(t, err)
		result, ok := keeper.GetIBCPath(ctx, chain.Name)
		assert.Equal(t, path, result)
		assert.True(t, ok)
	}).Repeat(repeats))

	t.Run("should return error when registered the same asset twice", testutils.Func(func(t *testing.T) {
		setup()
		path := randomIBCPath()
		chain := randomChain()
		chain.Assets = nil
		chain.IBCPath = ""
		keeper.SetCosmosChain(ctx, chain)
		err := keeper.RegisterIBCPath(ctx, chain.Name, path)
		assert.NoError(t, err)
		path2 := randomIBCPath()
		err2 := keeper.RegisterIBCPath(ctx, chain.Name, path2)
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
		axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		keeper = axelarnetKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace)
	}
	t.Run("should return list of registered cosmos chains", testutils.Func(func(t *testing.T) {
		setup()

		count := rand.I64Between(10, 100)
		chains := make([]string, count)

		for i := 0; i < int(count); i++ {
			chains[i] = strings.ToLower(rand.NormalizedStr(10))
			keeper.SetCosmosChain(ctx, types.CosmosChain{
				Name:       chains[i],
				AddrPrefix: rand.NormalizedStr(5),
			})
			keeper.RegisterAssetToCosmosChain(ctx, randomAsset(), chains[i])
		}
		sort.Strings(chains)
		assert.Equal(t, chains, keeper.GetCosmosChains(ctx))

	}).Repeat(repeats))

	t.Run("should empty list when no chain registered", testutils.Func(func(t *testing.T) {
		setup()
		empty := make([]string, 0)

		assert.Equal(t, empty, keeper.GetCosmosChains(ctx))

	}).Repeat(repeats))

}

func randomIBCPath() string {
	port := rand.NormalizedStrBetween(5, 10)
	identifier := fmt.Sprintf("%s%d", "channel-", rand.I64Between(0, 9999))
	return port + "/" + identifier
}

func randomAsset() string {
	return rand.Denom(5, 20)
}
