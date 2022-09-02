package keeper

// import (
// 	"testing"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/axelarnetwork/axelar-core/testutils/rand"
// 	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
// 	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
// 	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
// )

// // func TestGetMigrationHandler_deregisterUaxlAsset(t *testing.T) {
// // 	ctx, keeper := setup()
// // 	uaxlAsset := "uaxl"
// // 	anotherAsset := rand.NormalizedStr(5)

// // 	keeper.SetChain(ctx, axelarnet.Axelarnet)
// // 	keeper.SetChain(ctx, evm.Ethereum)
// // 	if err := keeper.RegisterAsset(ctx, axelarnet.Axelarnet, exported.NewAsset(uaxlAsset, true)); err != nil {
// // 		panic(err)
// // 	}
// // 	if err := keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(uaxlAsset, false)); err != nil {
// // 		panic(err)
// // 	}
// // 	if err := keeper.RegisterAsset(ctx, evm.Ethereum, exported.NewAsset(anotherAsset, false)); err != nil {
// // 		panic(err)
// // 	}

// // 	assert.True(t, keeper.IsAssetRegistered(ctx, axelarnet.Axelarnet, uaxlAsset))
// // 	assert.True(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, uaxlAsset))
// // 	assert.True(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, anotherAsset))

// // 	handler := GetMigrationHandler(keeper)
// // 	err := handler(ctx)
// // 	assert.NoError(t, err)

// // 	assert.True(t, keeper.IsAssetRegistered(ctx, axelarnet.Axelarnet, uaxlAsset))
// // 	assert.False(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, uaxlAsset))
// // 	assert.True(t, keeper.IsAssetRegistered(ctx, evm.Ethereum, anotherAsset))
// // }
