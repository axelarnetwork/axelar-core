package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmMock "github.com/axelarnetwork/axelar-core/x/evm/types/mock"
	"github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/legacy"
	legacyExported "github.com/axelarnetwork/axelar-core/x/nexus/legacy/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
)

func TestGetMigrationHandler_migrateCosmosChainState(t *testing.T) {
	ctx, keeper := setup()

	// set old cosmos chains
	oldCosmosChains := randCosmosChains()
	chainMap := map[string]axelarnetTypes.CosmosChain{}
	for _, c := range oldCosmosChains {
		chainMap[c.Name] = c
	}

	axelarnetKeeper := mock.AxelarnetKeeperMock{
		GetCosmosChainByNameFunc: func(ctx sdk.Context, chain string) (axelarnetTypes.CosmosChain, bool) {
			val, ok := chainMap[chain]
			return val, ok
		},
		GetCosmosChainsFunc: func(ctx sdk.Context) []string {
			names := make([]string, len(oldCosmosChains))
			for i, c := range oldCosmosChains {
				names[i] = c.Name
			}
			return names
		},
	}

	// set old chain state
	oldChainStates := make([]legacy.ChainState, len(oldCosmosChains))
	for i := 0; i < len(oldChainStates); i++ {
		assets := make([]string, len(oldCosmosChains[i].Assets))
		for j := 0; j < len(oldCosmosChains[i].Assets); j++ {
			assets[j] = oldCosmosChains[i].Assets[j].Denom
		}
		oldChainStates[i] = legacy.ChainState{
			Chain: legacyExported.Chain{
				Name:                  oldCosmosChains[i].Name,
				NativeAsset:           rand.Denom(5, 10),
				SupportsForeignAssets: true,
				Module:                rand.StrBetween(5, 10),
			},
			Activated: rand.Bools(0.5).Next(),
			Assets:    assets,
		}
	}

	// set legacy chain and states
	for _, state := range oldChainStates {
		keeper.getStore(ctx).Set(chainPrefix.Append(utils.LowerCaseKey(state.Chain.Name)), &state.Chain)
		keeper.getStore(ctx).Set(chainStatePrefix.Append(utils.LowerCaseKey(state.Chain.Name)), &state)
	}

	// should get old chain states
	for _, state := range oldChainStates {
		assert.NotNil(t, keeper.getStore(ctx).GetRaw(chainPrefix.Append(utils.LowerCaseKey(state.Chain.Name))))
		assert.NotNil(t, keeper.getStore(ctx).GetRaw(chainStatePrefix.Append(utils.LowerCaseKey(state.Chain.Name))))

	}

	handler := GetMigrationHandler(keeper, &axelarnetKeeper, &mock.EVMBaseKeeperMock{})
	err := handler(ctx)
	assert.NoError(t, err)

	// should migrate to new chain state
	for i, oldState := range oldChainStates {
		newChain := exported.Chain{
			Name:                  oldState.Chain.Name,
			SupportsForeignAssets: oldState.Chain.SupportsForeignAssets,
			KeyType:               oldState.Chain.KeyType,
			Module:                oldState.Chain.Module,
		}
		// get new chain state
		var newState types.ChainState
		newState, ok := keeper.getChainState(ctx, newChain)
		assert.True(t, ok)

		assert.Equal(t, oldState.Chain.Name, newState.Chain.Name)
		assert.Equal(t, oldState.Chain.SupportsForeignAssets, newState.Chain.SupportsForeignAssets)
		assert.Equal(t, oldState.Chain.Module, newState.Chain.Module)

		assert.Equal(t, oldState.Activated, newState.Activated)
		assert.Equal(t, oldState.Maintainers, newState.Maintainers)

		oldAssets := oldCosmosChains[i].Assets
		for j, asset := range newState.Assets {
			assert.Equal(t, oldAssets[j].Denom, asset.Denom)
			assert.Equal(t, oldAssets[j].MinAmount, asset.MinAmount)
			assert.True(t, asset.IsNativeAsset)
		}
	}
}

func TestGetMigrationHandler_migrateEVMChainState(t *testing.T) {
	ctx, keeper := setup()

	// erc20 token metadata
	metaData := testutils.RandomTokens()

	axelarnetKeeper := mock.AxelarnetKeeperMock{
		GetCosmosChainByNameFunc: func(ctx sdk.Context, chain string) (axelarnetTypes.CosmosChain, bool) { return axelarnetTypes.CosmosChain{}, true },
		GetCosmosChainsFunc:      func(ctx sdk.Context) []string { return nil },
	}

	evmChainKeeper := &evmMock.ChainKeeperMock{
		GetTokensFunc: func(ctx sdk.Context) []evmtypes.ERC20Token {
			tokens := make([]evmtypes.ERC20Token, len(metaData))
			for i, data := range metaData {
				token := evmtypes.CreateERC20Token(func(_ evmtypes.ERC20TokenMetadata) {}, data)
				tokens[i] = token
			}
			return tokens
		},
	}

	evmBaseKeeper := evmMock.BaseKeeperMock{
		ForChainFunc: func(string) evmtypes.ChainKeeper { return evmChainKeeper },
	}

	// set legacy chain and states
	evmChains := []string{"ethereum", "fantom", "avalanche", "moonbeam"}
	// set old chain state
	oldChainStates := make([]legacy.ChainState, len(evmChains))
	for i, name := range evmChains {
		chain := legacyExported.Chain{
			Name:                  name,
			SupportsForeignAssets: true,
			Module:                "evm",
			NativeAsset:           name + "-wei",
		}

		assets := make([]string, len(metaData)+1)
		assets[0] = name + "-wei"

		for j := 0; j < len(metaData); j++ {
			assets[j+1] = metaData[j].Asset
		}
		oldChainStates[i] = legacy.ChainState{
			Chain:       chain,
			Activated:   rand.Bools(0.5).Next(),
			Assets:      assets,
			Maintainers: randMaintainers(),
		}

		keeper.getStore(ctx).Set(chainPrefix.Append(utils.LowerCaseKey(name)), &chain)
		keeper.getStore(ctx).Set(chainStatePrefix.Append(utils.LowerCaseKey(name)), &oldChainStates[i])

	}

	// should get old chain states
	for _, state := range oldChainStates {
		assert.NotNil(t, keeper.getStore(ctx).GetRaw(chainPrefix.Append(utils.LowerCaseKey(state.Chain.Name))))
		assert.NotNil(t, keeper.getStore(ctx).GetRaw(chainStatePrefix.Append(utils.LowerCaseKey(state.Chain.Name))))
	}

	handler := GetMigrationHandler(keeper, &axelarnetKeeper, &evmBaseKeeper)
	err := handler(ctx)
	assert.NoError(t, err)

	// should migrate to new chain state
	for _, oldState := range oldChainStates {
		newChain := legacyToNewChain(oldState.Chain)
		// get new chain state
		var newState types.ChainState
		newState, ok := keeper.getChainState(ctx, newChain)
		assert.True(t, ok)

		assert.Equal(t, oldState.Chain.Name, newState.Chain.Name)
		assert.Equal(t, oldState.Chain.SupportsForeignAssets, newState.Chain.SupportsForeignAssets)
		assert.Equal(t, oldState.Chain.Module, newState.Chain.Module)

		assert.Equal(t, oldState.Activated, newState.Activated)
		assert.Equal(t, oldState.Maintainers, newState.Maintainers)

		for j, asset := range newState.Assets {
			assert.Equal(t, metaData[j].Asset, asset.Denom)
			assert.Equal(t, metaData[j].MinAmount, asset.MinAmount)
			assert.False(t, asset.IsNativeAsset)
		}
	}
}

func randCosmosChains() []axelarnetTypes.CosmosChain {
	chains := make([]axelarnetTypes.CosmosChain, rand.I64Between(10, 50))
	for i := 0; i < len(chains); i++ {
		assets := make([]axelarnetTypes.Asset, rand.I64Between(5, 10))
		for j := 0; j < len(assets); j++ {
			assets[j] = axelarnetTypes.Asset{
				Denom:     rand.Denom(10, 20),
				MinAmount: sdk.NewInt(rand.I64Between(100000, 1000000)),
			}
		}
		chains[i] = axelarnetTypes.CosmosChain{
			Name:       rand.StrBetween(5, 10),
			IBCPath:    rand.StrBetween(5, 10),
			AddrPrefix: rand.StrBetween(5, 10),
			Assets:     assets,
		}
	}

	return chains
}

func randMaintainers() []sdk.ValAddress {
	maintainers := make([]sdk.ValAddress, rand.I64Between(50, 100))
	for i := 0; i < len(maintainers); i++ {
		maintainers[i] = rand.ValAddr()
	}
	return maintainers
}

func legacyToNewChain(chain legacyExported.Chain) exported.Chain {
	return exported.Chain{
		Name:                  chain.Name,
		SupportsForeignAssets: chain.SupportsForeignAssets,
		KeyType:               chain.KeyType,
		Module:                chain.Module,
	}
}
