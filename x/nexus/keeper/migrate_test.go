package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/slices"
)

func TestGetMigrationHandler_deregisterUaxlAsset(t *testing.T) {
	ctx, keeper := setup()
	uaxlAsset := "uaxl"
	anotherAsset := rand.NormalizedStr(5)

	keeper.SetChain(ctx, axelarnet.Axelarnet)
	keeper.SetChain(ctx, evm.Ethereum)
	if err := keeper.RegisterAsset(ctx, axelarnet.Axelarnet, exported.NewAsset(uaxlAsset, true)); err != nil {
		panic(err)
	}
	if err := keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(uaxlAsset, false)); err != nil {
		panic(err)
	}
	if err := keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(anotherAsset, false)); err != nil {
		panic(err)
	}

	assert.True(t, keeper.IsAssetRegistered(ctx, axelarnet.Axelarnet, uaxlAsset))
	assert.True(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, uaxlAsset))
	assert.True(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, anotherAsset))

	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)

	assert.True(t, keeper.IsAssetRegistered(ctx, axelarnet.Axelarnet, uaxlAsset))
	assert.False(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, uaxlAsset))
	assert.True(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, anotherAsset))
}

func TestGetMigrationHandler_removeUaxlFeeInfo(t *testing.T) {
	ctx, keeper := setup()
	uaxlAsset := "uaxl"
	anotherAsset := rand.NormalizedStr(5)

	keeper.SetChain(ctx, axelarnet.Axelarnet)
	keeper.SetChain(ctx, evm.Ethereum)
	if err := keeper.RegisterAsset(ctx, axelarnet.Axelarnet, exported.NewAsset(uaxlAsset, true)); err != nil {
		panic(err)
	}
	if err := keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(uaxlAsset, false)); err != nil {
		panic(err)
	}
	if err := keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(anotherAsset, false)); err != nil {
		panic(err)
	}

	expectedFeeInfo := exported.NewFeeInfo(axelarnet.Axelarnet.Name, uaxlAsset, sdk.OneDec(), sdk.NewInt(100), sdk.NewInt(100000))
	expectedEthereumFeeInfo := exported.NewFeeInfo(evm.Ethereum.Name, uaxlAsset, sdk.OneDec(), sdk.NewInt(100), sdk.NewInt(100000))
	anotherAssetFeeInfo := exported.NewFeeInfo(axelarnet.Axelarnet.Name, anotherAsset, sdk.OneDec(), sdk.NewInt(10), sdk.NewInt(10000))

	if err := keeper.RegisterFee(ctx, axelarnet.Axelarnet, expectedFeeInfo); err != nil {
		panic(err)
	}
	if err := keeper.RegisterFee(ctx, evm.Ethereum, expectedEthereumFeeInfo); err != nil {
		panic(err)
	}
	if err := keeper.RegisterFee(ctx, evm.Ethereum, anotherAssetFeeInfo); err != nil {
		panic(err)
	}

	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)

	actualFeeInfo, ok := keeper.GetFeeInfo(ctx, axelarnet.Axelarnet, uaxlAsset)
	assert.True(t, ok)
	assert.Equal(t, expectedFeeInfo, actualFeeInfo)

	actualFeeInfo, ok = keeper.GetFeeInfo(ctx, evm.Ethereum, uaxlAsset)
	assert.False(t, ok)

	actualFeeInfo, ok = keeper.GetFeeInfo(ctx, evm.Ethereum, anotherAsset)
	assert.True(t, ok)
	assert.Equal(t, anotherAssetFeeInfo, actualFeeInfo)
}

func TestGetMigrationHandler_addNewParams(t *testing.T) {
	ctx, keeper := setup()

	defaultParams := types.DefaultParams()
	keeper.params.Set(ctx, types.KeyChainActivationThreshold, defaultParams.ChainActivationThreshold)
	assert.Panics(t, func() { keeper.GetParams(ctx) })

	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, defaultParams, keeper.GetParams(ctx))
}

func TestGetMigrationHandler_migrateChainMaintainers(t *testing.T) {
	ctx, keeper := setup()

	maintainersCount := rand.I64Between(1, 2)
	maintainers := make([]sdk.ValAddress, maintainersCount)
	for i := 0; i < int(maintainersCount); i++ {
		maintainers[i] = rand.ValAddr()
	}
	chainState := types.ChainState{
		Chain:       exported.Chain{Name: exported.ChainName(rand.NormalizedStr(5))},
		Maintainers: maintainers,
	}
	keeper.SetChainState(ctx, &chainState)

	handler := GetMigrationHandler(keeper)
	err := handler(ctx)
	assert.NoError(t, err)

	actual := keeper.getChainStates(ctx)
	assert.Len(t, actual, 1)
	assert.Equal(t, chainState.Chain, actual[0].Chain)
	assert.Equal(t, chainState.Maintainers, actual[0].Maintainers)
	assert.ElementsMatch(t, maintainers, slices.Map(actual[0].MaintainerStates, func(mt types.MaintainerState) sdk.ValAddress { return mt.Address }))
}
