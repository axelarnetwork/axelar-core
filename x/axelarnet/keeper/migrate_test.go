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
)

func TestStoreMigration(t *testing.T) {

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
