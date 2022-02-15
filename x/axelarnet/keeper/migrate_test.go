package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/legacy"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
)

func TestGetMigrationHandler_rebuildCosmosChain(t *testing.T) {
	encCfg := appParams.MakeEncodingConfig()
	axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace)

	// create rand cosmos chains
	randChains := make([]types.CosmosChain, rand.I64Between(10, 50))
	for i := 0; i < len(randChains); i++ {
		chain := randomChain()

		// mock the state
		// set chain
		keeper.SetCosmosChain(ctx, types.CosmosChain{
			Name:       chain.Name,
			AddrPrefix: chain.AddrPrefix,
		})
		// set path
		keeper.getStore(ctx).SetRaw(legacy.PathPrefix.Append(utils.LowerCaseKey(chain.Name)), []byte(chain.IBCPath))
		for _, asset := range chain.Assets {
			// set chain by asset
			keeper.getStore(ctx).SetRaw(legacy.ChainByAssetPrefix.
				Append(utils.LowerCaseKey(asset.Denom)), []byte(chain.Name))

			// set asset by chain
			keeper.getStore(ctx).Set(legacy.AssetByChainPrefix.
				Append(utils.LowerCaseKey(chain.Name)).
				Append(utils.LowerCaseKey(asset.Denom)), &asset)
		}

		randChains[i] = chain
	}

	for _, chain := range randChains {
		// should get cosmos chain
		c, ok := keeper.GetCosmosChainByName(ctx, chain.Name)
		assert.True(t, ok)
		assert.Equal(t, types.CosmosChain{Name: chain.Name, AddrPrefix: chain.AddrPrefix}, c)

		// should get path
		bz := keeper.getStore(ctx).GetRaw(legacy.PathPrefix.Append(utils.LowerCaseKey(chain.Name)))
		assert.NotNil(t, bz)
		assert.ElementsMatch(t, chain.Assets, getAssets(ctx, keeper, chain.Name))
		for _, asset := range chain.Assets {
			// should get chain by asset
			bz := keeper.getStore(ctx).GetRaw(legacy.ChainByAssetPrefix.Append(utils.LowerCaseKey(asset.Denom)))
			assert.NotNil(t, bz)
		}
	}

	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)

	for _, chain := range randChains {
		// should get cosmos chain
		updated, ok := keeper.GetCosmosChainByName(ctx, chain.Name)
		assert.True(t, ok)
		assert.Equal(t, chain.Name, updated.Name)
		assert.Equal(t, chain.IBCPath, updated.IBCPath)
		assert.Equal(t, chain.AddrPrefix, updated.AddrPrefix)
		assert.ElementsMatch(t, chain.Assets, updated.Assets)
	}

}

func TestGetMigrationHandler_deletePrefix(t *testing.T) {

	encCfg := appParams.MakeEncodingConfig()
	axelarnetSubspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	keeper := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), axelarnetSubspace)

	// set path
	randChains := make([]string, rand.I64Between(10, 50))
	for i := 0; i < len(randChains); i++ {
		randChains[i] = rand.StrBetween(5, 10)
	}

	for _, chain := range randChains {
		key := legacy.PathPrefix.Append(utils.LowerCaseKey(chain))
		keeper.getStore(ctx).SetRaw(key, []byte(rand.Str(5)+"/"+rand.Str(5)))
	}

	// should get path
	for _, chain := range randChains {
		key := legacy.PathPrefix.Append(utils.LowerCaseKey(chain))
		bz := keeper.getStore(ctx).GetRaw(key)
		assert.NotNil(t, bz)
	}

	// set chain by asset
	randAssets := make([]string, len(randChains))
	for i := 0; i < len(randChains); i++ {
		randAssets[i] = rand.StrBetween(5, 10)
	}

	for i, asset := range randAssets {
		key := legacy.ChainByAssetPrefix.Append(utils.LowerCaseKey(asset))
		keeper.getStore(ctx).SetRaw(key, []byte(randChains[i]))
	}

	// should get chain by asset
	for _, asset := range randAssets {
		key := legacy.ChainByAssetPrefix.Append(utils.LowerCaseKey(asset))
		bz := keeper.getStore(ctx).GetRaw(key)
		assert.NotNil(t, bz)
	}

	// set asset by chain
	for i, chain := range randChains {
		key := legacy.AssetByChainPrefix.Append(utils.LowerCaseKey(chain))
		keeper.getStore(ctx).SetRaw(key, []byte(randAssets[i]))
	}

	// should get asset by chain
	for _, chain := range randChains {
		key := legacy.AssetByChainPrefix.Append(utils.LowerCaseKey(chain))
		bz := keeper.getStore(ctx).GetRaw(key)
		assert.NotNil(t, bz)
	}

	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)

	// should not get path
	for _, chain := range randChains {
		key := legacy.PathPrefix.Append(utils.LowerCaseKey(chain))
		bz := keeper.getStore(ctx).GetRaw(key)
		assert.Nil(t, bz)
	}

	// should not get chain by asset
	for _, asset := range randAssets {
		key := legacy.ChainByAssetPrefix.Append(utils.LowerCaseKey(asset))
		bz := keeper.getStore(ctx).GetRaw(key)
		assert.Nil(t, bz)
	}

	// should not get chain by asset
	for _, chain := range randChains {
		key := legacy.AssetByChainPrefix.Append(utils.LowerCaseKey(chain))
		bz := keeper.getStore(ctx).GetRaw(key)
		assert.Nil(t, bz)
	}
}

func randomChain() types.CosmosChain {
	assets := make([]types.Asset, rand.I64Between(10, 50))
	for i := 0; i < len(assets); i++ {
		assets[i] = randAsset()
	}
	return types.CosmosChain{
		Name:       rand.Denom(5, 10),
		IBCPath:    rand.Str(5) + "/" + rand.Str(5),
		AddrPrefix: rand.Denom(5, 10),
		Assets:     assets,
	}
}

func randAsset() types.Asset {
	return types.Asset{
		Denom:     rand.Denom(5, 10),
		MinAmount: sdk.NewInt(rand.I64Between(100000, 1000000)),
	}

}
