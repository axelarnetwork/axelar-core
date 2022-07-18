package keeper

import (
	mathrand "math/rand"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types/mock"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
	rand2 "github.com/axelarnetwork/utils/test/rand"
)

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx            sdk.Context
		paramStoreKey  sdk.StoreKey
		paramTStoreKey sdk.StoreKey
		paramstore     prefix.Store
		k              Keeper
		expectedParams types.Params
	)

	Given("a context", func() {
		store := fake.NewMultiStore()
		ctx = sdk.NewContext(store, tmproto.Header{}, false, log.TestingLogger())

	}).
		Given("a param store with the old tss params", func() {
			paramStoreKey = sdk.NewKVStoreKey(paramstypes.StoreKey)
			paramTStoreKey = sdk.NewKVStoreKey(paramstypes.TStoreKey)

			paramstore = prefix.NewStore(ctx.MultiStore().GetKVStore(paramStoreKey), append([]byte(types.StoreKey), '/'))

		}).
		Given("a keeper with access to that store", func() {
			encodingConfig := params.MakeEncodingConfig()
			subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, paramStoreKey, paramTStoreKey, types.StoreKey)
			banker := mock.BankerMock{}
			distributor := mock.DistributorMock{}
			staker := mock.StakerMock{}
			k = NewKeeper(encodingConfig.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace, &banker, &distributor, &staker)
		}).
		When("the pre-migration params are set up in the param store", func() {
			expectedParams = types.Params{
				ExternalChainVotingInflationRate: sdk.MustNewDecFromStr(strconv.FormatFloat(mathrand.Float64(), 'f', 3, 64)),
				KeyMgmtRelativeInflationRate:     sdk.MustNewDecFromStr(strconv.FormatFloat(mathrand.Float64(), 'f', 3, 64)),
			}

			paramstore.Set(types.KeyExternalChainVotingInflationRate, funcs.Must(types.ModuleCdc.LegacyAmino.MarshalJSON(expectedParams.ExternalChainVotingInflationRate)))
			paramstore.Set(keyTssRelativeInflationRate, funcs.Must(types.ModuleCdc.LegacyAmino.MarshalJSON(expectedParams.KeyMgmtRelativeInflationRate)))

			// Params with the post-migration keeper reads different keys, so it shouldn't be able to fill the params correctly
			assert.NotEqual(t, expectedParams, k.GetParams(ctx))
		}).
		When("a tss reward pool is set up", func() {
			tssPool := k.GetPool(ctx, tss.ModuleName)
			tssPool.AddReward(rand.ValAddr(), sdk.NewInt64Coin(rand2.AlphaStrBetween(5, 10), rand.PosI64()))

			_, ok := k.getPool(ctx, multisig.ModuleName)
			assert.False(t, ok)
		}).
		Then("params are migrated", func(t *testing.T) {
			handler := GetMigrationHandler(k, paramStoreKey, paramTStoreKey)
			assert.NoError(t, handler(ctx))

			assert.Nil(t, paramstore.Get(keyTssRelativeInflationRate))

			actualParams := k.GetParams(ctx)
			assert.Equal(t, expectedParams, actualParams)
		}).
		Then("the pool is migrated", func(t *testing.T) {
			_, ok := k.getPool(ctx, tss.ModuleName)
			assert.False(t, ok)

			_, ok = k.getPool(ctx, multisig.ModuleName)
			assert.True(t, ok)
		})

}
