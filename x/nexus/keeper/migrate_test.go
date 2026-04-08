package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestMigrate8to9(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("nexusKey"), store.NewKVStoreKey("tNexusKey"), "nexus")
	k := keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("nexus"), subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))

	chain := testutils.RandomChain()
	addr := rand.ValAddr()
	oldMaxSize := 1 << 15 // 32,768
	newMaxSize := types.MaxBitmapSize()
	voteCount := newMaxSize + 500

	Given("a chain with a maintainer that has an old large bitmap", func() {
		k.SetChain(ctx, chain)
		assert.NoError(t, k.AddChainMaintainer(ctx, chain, addr))

		ms, ok := k.GetChainMaintainerState(ctx, chain, addr)
		assert.True(t, ok)

		maintainer := ms.(*types.MaintainerState)
		maintainer.MissingVotes.TrueCountCache.SetMaxSize(oldMaxSize)
		maintainer.IncorrectVotes.TrueCountCache.SetMaxSize(oldMaxSize)

		for i := 0; i < voteCount; i++ {
			maintainer.MarkMissingVote(i%3 == 0)
			maintainer.MarkIncorrectVote(i%5 == 0)
		}

		assert.Greater(t, len(maintainer.MissingVotes.TrueCountCache.CumulativeValue), newMaxSize)
		assert.NoError(t, k.SetChainMaintainerState(ctx, maintainer))
	}).
		When("the migration runs", func() {
			assert.NoError(t, keeper.Migrate8to9(k)(ctx))
		}).
		Then("bitmap MaxSize should be capped and buffer shrinks on next Add", func(t *testing.T) {
			ms, ok := k.GetChainMaintainerState(ctx, chain, addr)
			assert.True(t, ok)

			maintainer := ms.(*types.MaintainerState)
			assert.Equal(t, int32(newMaxSize), maintainer.MissingVotes.TrueCountCache.MaxSize)
			assert.Equal(t, int32(newMaxSize), maintainer.IncorrectVotes.TrueCountCache.MaxSize)

			// The buffer is still large in storage until the next Add triggers shrink
			assert.Greater(t, len(maintainer.MissingVotes.TrueCountCache.CumulativeValue), newMaxSize)

			// After one more vote, shrink should trigger
			maintainer.MarkMissingVote(false)
			assert.Equal(t, newMaxSize, len(maintainer.MissingVotes.TrueCountCache.CumulativeValue))

			// Vote counts within a small window should still be correct after shrink.
			// The window of 100 covers the 1 extra false vote plus the last 99 loop
			// iterations (i=1425..1523). Votes where i%3==0 are true: 1425,1428,...,1521 = 33.
			missingCount := maintainer.CountMissingVotes(100)
			assert.Equal(t, uint64(33), missingCount)
		}).
		Run(t)
}

func TestMigrate6to7(t *testing.T) {
	encCfg := app.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, store.NewKVStoreKey("nexusKey"), store.NewKVStoreKey("tNexusKey"), "nexus")
	k := keeper.NewKeeper(encCfg.Codec, store.NewKVStoreKey("nexus"), subspace)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))

	Given("subspace is setup with params before migration", func() {
		subspace.Set(ctx, types.KeyChainActivationThreshold, types.DefaultParams().ChainActivationThreshold)
		subspace.Set(ctx, types.KeyChainMaintainerMissingVoteThreshold, types.DefaultParams().ChainMaintainerMissingVoteThreshold)
		subspace.Set(ctx, types.KeyChainMaintainerIncorrectVoteThreshold, types.DefaultParams().ChainMaintainerIncorrectVoteThreshold)
		subspace.Set(ctx, types.KeyChainMaintainerCheckWindow, types.DefaultParams().ChainMaintainerCheckWindow)
	}).
		When("", func() {}).
		Then("the migration should add the new param with the default value", func(t *testing.T) {
			actualGateway := sdk.AccAddress{}
			actualEndBlockerLimit := uint64(0)

			assert.PanicsWithError(t, "UnmarshalJSON cannot decode empty bytes", func() {
				subspace.Get(ctx, types.KeyGateway, &actualGateway)
			})
			assert.PanicsWithError(t, "UnmarshalJSON cannot decode empty bytes", func() {
				subspace.Get(ctx, types.KeyEndBlockerLimit, &actualEndBlockerLimit)
			})
			assert.PanicsWithError(t, "UnmarshalJSON cannot decode empty bytes", func() {
				k.GetParams(ctx)
			})

			keeper.Migrate6to7(k)(ctx)

			assert.NotPanics(t, func() {
				subspace.Get(ctx, types.KeyGateway, &actualGateway)
			})
			assert.NotPanics(t, func() {
				subspace.Get(ctx, types.KeyEndBlockerLimit, &actualEndBlockerLimit)
			})
			assert.NotPanics(t, func() {
				k.GetParams(ctx)
			})

			assert.Equal(t, types.DefaultParams().Gateway, actualGateway)
			assert.Equal(t, types.DefaultParams().Gateway, k.GetParams(ctx).Gateway)
			assert.Equal(t, types.DefaultParams().EndBlockerLimit, actualEndBlockerLimit)
			assert.Equal(t, types.DefaultParams().EndBlockerLimit, k.GetParams(ctx).EndBlockerLimit)
		}).
		Run(t)

}
