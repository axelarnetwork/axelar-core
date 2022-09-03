package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

func TestGetMigrationHandler(t *testing.T) {
	ctx, keeper := setup()

	t.Run("migrateMaintainerStates", func(t *testing.T) {
		chain := testutils.Chain()
		window := int(types.DefaultParams().ChainMaintainerCheckWindow)
		boolGen := rand.Bools(0.5)
		chainState := types.ChainState{
			Chain: chain,
			MaintainerStates: slices.Expand(func(int) types.MaintainerState {
				maintainerState := types.NewMaintainerState("", rand.ValAddr())
				for i := 0; i < window; i++ {
					maintainerState.MarkIncorrectVote(boolGen.Next())
					maintainerState.MarkMissingVote(boolGen.Next())
				}

				return *maintainerState
			}, 5),
		}
		keeper.setChainState(ctx, chainState)

		migrateHandler := GetMigrationHandler(keeper)
		migrateHandler(ctx)

		actual := funcs.MustOk(keeper.getChainState(ctx, chain))
		assert.Nil(t, actual.MaintainerStates)

		for _, maintainerState := range chainState.MaintainerStates {
			actual := funcs.MustOk(keeper.getChainMaintainerState(ctx, chain.Name, maintainerState.Address))

			assert.Equal(t, chain.Name, actual.Chain)
			assert.Equal(t, maintainerState.CountIncorrectVotes(window), actual.CountIncorrectVotes(window))
			assert.Equal(t, maintainerState.CountMissingVotes(window), actual.CountMissingVotes(window))
		}
	})
}
