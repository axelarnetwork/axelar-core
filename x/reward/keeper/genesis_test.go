package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
)

func TestExportGenesis(t *testing.T) {
	ctx, keeper, _, _, _ := setup()
	keeper.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), []types.Pool{}))

	poolName1 := "aaaaa"
	pool1 := keeper.GetPool(ctx, poolName1)

	pool1.AddReward(rand.ValAddr(), sdk.NewCoin(denom, sdk.ZeroInt()))

	validator2 := rand.ValAddr()
	coin2 := sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000)))
	pool1.AddReward(validator2, coin2)

	validator3 := rand.ValAddr()
	pool1.AddReward(validator3, sdk.NewCoin(denom, sdk.ZeroInt()))
	pool1.ClearRewards(validator3)

	poolName2 := "bbbbb"
	pool2 := keeper.GetPool(ctx, poolName2)

	validator4 := rand.ValAddr()
	coin4 := sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000)))
	pool2.AddReward(validator4, coin4)

	expectedPool1 := types.NewPool(poolName1)
	expectedPool1.Rewards = []types.Pool_Reward{
		{
			Validator: validator2,
			Coins:     sdk.NewCoins(coin2),
		},
	}
	expectedPool2 := types.NewPool(poolName2)
	expectedPool2.Rewards = []types.Pool_Reward{
		{
			Validator: validator4,
			Coins:     sdk.NewCoins(coin4),
		},
	}
	expected := types.NewGenesisState(
		types.DefaultParams(),
		[]types.Pool{expectedPool1, expectedPool2},
	)
	actual := keeper.ExportGenesis(ctx)

	assert.Equal(t, expected, actual)
	assert.NoError(t, actual.Validate())
}

func TestInitGenesis(t *testing.T) {
	ctx, keeper, _, _, _ := setup()

	expectedPool1 := types.NewPool("aaaaa")
	expectedPool1.Rewards = []types.Pool_Reward{
		{
			Validator: rand.ValAddr(),
			Coins:     sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000)))),
		},
		{
			Validator: rand.ValAddr(),
			Coins:     sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000)))),
		},
		{
			Validator: rand.ValAddr(),
			Coins:     sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000)))),
		},
	}
	expectedPool2 := types.NewPool("bbbbb")
	expectedPool2.Rewards = []types.Pool_Reward{
		{
			Validator: rand.ValAddr(),
			Coins:     sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000)))),
		},
	}
	expectedPool3 := types.NewPool("ccccc")
	expectedPool3.Rewards = nil
	expected := types.NewGenesisState(
		types.DefaultParams(),
		[]types.Pool{expectedPool1, expectedPool2, expectedPool3},
	)

	keeper.InitGenesis(ctx, expected)
	actual := keeper.ExportGenesis(ctx)

	assert.Equal(t, expected, actual)
	assert.NoError(t, actual.Validate())
}
