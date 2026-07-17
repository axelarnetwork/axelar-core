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
		Then("bitmap MaxSize and buffer should be shrunk eagerly", func(t *testing.T) {
			ms, ok := k.GetChainMaintainerState(ctx, chain, addr)
			assert.True(t, ok)

			maintainer := ms.(*types.MaintainerState)
			assert.Equal(t, int32(newMaxSize), maintainer.MissingVotes.TrueCountCache.MaxSize)
			assert.Equal(t, int32(newMaxSize), maintainer.IncorrectVotes.TrueCountCache.MaxSize)

			// The migration reallocates the buffer immediately; no further Add is
			// needed, so maintainers on deactivated chains stop carrying oversized
			// buffers in storage.
			assert.Equal(t, newMaxSize, len(maintainer.MissingVotes.TrueCountCache.CumulativeValue))
			assert.Equal(t, newMaxSize, len(maintainer.IncorrectVotes.TrueCountCache.CumulativeValue))

			// Vote counts within a window smaller than the cap remain correct after
			// the shrink. The most recent 100 votes are loop iterations i=1424..1523;
			// those where i%3==0 are missing votes: 1425,1428,...,1521 = 33.
			missingCount := maintainer.CountMissingVotes(100)
			assert.Equal(t, uint64(33), missingCount)
		}).
		Run(t)
}
