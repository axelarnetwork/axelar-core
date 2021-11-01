package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types/mock"
)

const denom = "test"

func setup() (sdk.Context, exported.RewardPool, Keeper, *mock.BankerMock, *mock.DistributorMock, *mock.StakerMock) {
	banker := mock.BankerMock{}
	distributor := mock.DistributorMock{}
	staker := mock.StakerMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encodingConfig := params.MakeEncodingConfig()
	subspace := paramstypes.NewSubspace(encodingConfig.Marshaler, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "reward")
	keeper := NewKeeper(encodingConfig.Marshaler, sdk.NewKVStoreKey(types.StoreKey), subspace, &banker, &distributor, &staker)

	return ctx, keeper.GetPool(ctx, rand.Str(10)), keeper, &banker, &distributor, &staker
}

func TestAddReward(t *testing.T) {
	ctx, pool, keeper, _, _, _ := setup()
	validator := rand.ValAddr()
	coin1 := sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000000)))

	pool.AddReward(validator, coin1)
	p := pool.(*rewardPool)

	assert.Len(t, p.Rewards, 1)
	assert.Len(t, p.Rewards[0].Coins, 1)
	assert.Equal(t, validator, p.Rewards[0].Validator)
	assert.Equal(t, coin1, p.Rewards[0].Coins[0])
	assert.Equal(t, p, keeper.GetPool(ctx, p.Name).(*rewardPool))

	coin2 := sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000000)))
	pool.AddReward(validator, coin2)
	p = pool.(*rewardPool)

	assert.Len(t, p.Rewards, 1)
	assert.Len(t, p.Rewards[0].Coins, 1)
	assert.Equal(t, validator, p.Rewards[0].Validator)
	assert.Equal(t, coin1.Add(coin2), p.Rewards[0].Coins[0])
	assert.Equal(t, p, keeper.GetPool(ctx, p.Name).(*rewardPool))
}

func TestReleaseRewards(t *testing.T) {
	ctx, pool, keeper, banker, distributor, staker := setup()
	validator := rand.ValAddr()
	coin := sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000000)))

	banker.MintCoinsFunc = func(ctx sdk.Context, name string, amt sdk.Coins) error { return nil }
	banker.SendCoinsFromModuleToModuleFunc = func(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error {
		return nil
	}
	staker.ValidatorFunc = func(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI { return stakingtypes.Validator{} }
	distributor.AllocateTokensToValidatorFunc = func(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) {}

	pool.AddReward(validator, coin)
	pool.ReleaseRewards(validator)
	p := pool.(*rewardPool)

	assert.Len(t, banker.MintCoinsCalls(), 1)
	assert.Len(t, banker.MintCoinsCalls()[0].Amt, 1)
	assert.Equal(t, coin, banker.MintCoinsCalls()[0].Amt[0])
	assert.Len(t, banker.SendCoinsFromModuleToModuleCalls(), 1)
	assert.Len(t, banker.SendCoinsFromModuleToModuleCalls()[0].Amt, 1)
	assert.Equal(t, coin, banker.SendCoinsFromModuleToModuleCalls()[0].Amt[0])
	assert.Len(t, staker.ValidatorCalls(), 1)
	assert.Equal(t, validator, staker.ValidatorCalls()[0].Addr)
	assert.Len(t, distributor.AllocateTokensToValidatorCalls(), 1)
	assert.Len(t, distributor.AllocateTokensToValidatorCalls()[0].Tokens, 1)
	assert.Equal(t, sdk.NewDecCoinFromCoin(coin), distributor.AllocateTokensToValidatorCalls()[0].Tokens[0])
	assert.Len(t, p.Rewards, 0)
	assert.Len(t, keeper.GetPool(ctx, p.Name).(*rewardPool).Rewards, 0)
}

func TestClearRewards(t *testing.T) {
	ctx, pool, keeper, _, _, _ := setup()
	validator := rand.ValAddr()
	coin := sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(10, 10000000)))

	pool.AddReward(validator, coin)
	pool.ClearRewards(validator)
	p := pool.(*rewardPool)

	assert.Len(t, p.Rewards, 0)
	assert.Len(t, keeper.GetPool(ctx, p.Name).(*rewardPool).Rewards, 0)

	pool.AddReward(validator, coin)
	pool.AddReward(rand.ValAddr(), coin)

	pool.ClearRewards(validator)
	assert.Len(t, p.Rewards, 1)
	assert.Len(t, keeper.GetPool(ctx, p.Name).(*rewardPool).Rewards, 1)
}
