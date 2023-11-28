package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	rewardKeeper "github.com/axelarnetwork/axelar-core/x/reward/keeper"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types/mock"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestKeeper_Inflation(t *testing.T) {
	var (
		k                      rewardKeeper.Keeper
		mintK                  *mock.MinterMock
		nexusK                 *mock.NexusMock
		q                      rewardKeeper.Querier
		ctx                    sdk.Context
		response               *types.InflationRateResponse
		err                    error
		paramsSubspace         paramstypes.Subspace
		tmInflation            sdk.Dec
		keyRelativeInflation   sdk.Dec
		externalChainInflation sdk.Dec
		chains                 []nexus.Chain
		activeStatus           map[nexus.ChainName]bool
		val                    sdk.ValAddress
	)

	given := Given("a reward keeper", func() {
		encCfg := app.MakeEncodingConfig()
		mintK = &mock.MinterMock{}
		nexusK = &mock.NexusMock{}
		store := fake.NewMultiStore()
		ctx = sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger())
		paramsSubspace = paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("rewardKey"), sdk.NewKVStoreKey("trewardKey"), "reward")

		k = rewardKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("reward"), paramsSubspace, nil, nil, nil)
		q = rewardKeeper.NewGRPCQuerier(k, mintK, nexusK)
	})

	whenParamsAreSet := When("params are set", func() {
		keyRelativeInflation = rand.ThresholdDec()
		externalChainInflation = rand.ThresholdDec()

		paramsSubspace.SetParamSet(ctx, &types.Params{
			KeyMgmtRelativeInflationRate:     keyRelativeInflation,
			ExternalChainVotingInflationRate: externalChainInflation,
		})

		tmInflation = rand.ThresholdDec()
		mintK.GetMinterFunc = func(ctx sdk.Context) minttypes.Minter {
			return minttypes.Minter{
				Inflation: tmInflation,
			}
		}
	})

	given.
		When2(whenParamsAreSet).
		When("one chain is active", func() {
			nexusK.GetChainMaintainersFunc = func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress { return []sdk.ValAddress{rand.ValAddr()} }
			nexusK.GetChainsFunc = func(ctx sdk.Context) []nexus.Chain {
				return []nexus.Chain{
					{Name: nexus.ChainName("test")},
					{Name: nexus.ChainName("test2")},
				}
			}
			nexusK.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return chain.Name == "test" }
		}).
		Then("query inflation", func(t *testing.T) {
			response, err = q.InflationRate(sdk.WrapSDKContext(ctx), &types.InflationRateRequest{})
			assert.NoError(t, err)

			keyManagementInflation := tmInflation.Mul(keyRelativeInflation)
			assert.Equal(t, response.InflationRate, tmInflation.Add(keyManagementInflation).Add(externalChainInflation))
		}).
		Run(t)

	given.
		When2(whenParamsAreSet).
		When("one chain is active", func() {
			val = rand.ValAddr()
			nexusK.GetChainMaintainersFunc = func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress { return []sdk.ValAddress{val} }
			nexusK.GetChainsFunc = func(ctx sdk.Context) []nexus.Chain {
				return []nexus.Chain{
					{Name: nexus.ChainName("test")},
				}
			}
			nexusK.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return true }
		}).
		Then("query inflation", func(t *testing.T) {
			response, err = q.InflationRate(sdk.WrapSDKContext(ctx), &types.InflationRateRequest{Validator: val.String()})
			assert.NoError(t, err)

			keyManagementInflation := tmInflation.Mul(keyRelativeInflation)
			assert.Equal(t, response.InflationRate, tmInflation.Add(keyManagementInflation).Add(externalChainInflation))
		}).
		Then("query inflation for a non chain maintainer", func(t *testing.T) {
			response, err = q.InflationRate(sdk.WrapSDKContext(ctx), &types.InflationRateRequest{Validator: rand.ValAddr().String()})
			assert.NoError(t, err)

			keyManagementInflation := tmInflation.Mul(keyRelativeInflation)
			assert.Equal(t, response.InflationRate, tmInflation.Add(keyManagementInflation))
		}).
		Run(t)

	given.
		When2(whenParamsAreSet).
		When("chains are set", func() {
			chains = slices.Expand(func(idx int) nexus.Chain {
				return nexus.Chain{Name: nexus.ChainName(rand.StrBetween(5, 10))}
			}, int(rand.I64Between(0, 10)))

			activeStatus = make(map[nexus.ChainName]bool)
			slices.ForEach(chains, func(chain nexus.Chain) {
				activeStatus[chain.Name] = rand.Bools(0.5).Next()
			})

			nexusK.IsChainActivatedFunc = func(ctx sdk.Context, chain nexus.Chain) bool { return activeStatus[chain.Name] }
			nexusK.GetChainsFunc = func(ctx sdk.Context) []nexus.Chain { return chains }
			nexusK.GetChainMaintainersFunc = func(ctx sdk.Context, chain nexus.Chain) []sdk.ValAddress { return []sdk.ValAddress{rand.ValAddr()} }
		}).
		Then("query inflation", func(t *testing.T) {
			response, err = q.InflationRate(sdk.WrapSDKContext(ctx), &types.InflationRateRequest{})
			assert.NoError(t, err)

			keyManagementInflation := tmInflation.Mul(keyRelativeInflation)
			externalVotingInflation := externalChainInflation.MulInt64(int64(len(slices.Filter(chains, func(chain nexus.Chain) bool {
				return activeStatus[chain.Name]
			}))))

			assert.Equal(t, response.InflationRate, tmInflation.Add(keyManagementInflation).Add(externalVotingInflation))
		}).
		Run(t, 10)
}
