package multisig_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/multisig/types/mock"
	typestestutils "github.com/axelarnetwork/axelar-core/x/multisig/types/testutils"
	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	rewardmock "github.com/axelarnetwork/axelar-core/x/reward/exported/mock"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
	. "github.com/axelarnetwork/utils/test"
)

func TestEndBlocker(t *testing.T) {
	var (
		ctx      sdk.Context
		k        *mock.KeeperMock
		rewarder *mock.RewarderMock
	)

	givenKeepersAndCtx := Given("keepers", func() {
		ctx = rand.Context(fake.NewMultiStore())
		k = &mock.KeeperMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
		}
		rewarder = &mock.RewarderMock{}
	})

	t.Run("handleKeygens", func(t *testing.T) {
		givenKeepersAndCtx.
			When("a pending keygen session expiry equal to the block height", func() {
				k.GetKeygenSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.KeygenSession {
					if expiry != ctx.BlockHeight() {
						return nil
					}

					return []types.KeygenSession{{
						Key: types.Key{
							ID:       testutils.KeyID(),
							Snapshot: snapshottestutils.Snapshot(uint64(rand.I64Between(10, 11)), utils.NewThreshold(1, 2)),
						},
						State: exported.Pending,
					}}
				}
			}).
			Then("should delete and penalize missing participants", func(t *testing.T) {
				pool := rewardmock.RewardPoolMock{
					ClearRewardsFunc: func(sdk.ValAddress) {},
				}

				k.DeleteKeygenSessionFunc = func(sdk.Context, exported.KeyID) {}
				rewarder.GetPoolFunc = func(sdk.Context, string) reward.RewardPool { return &pool }

				_, err := multisig.EndBlocker(ctx, abci.RequestEndBlock{}, k, rewarder)

				assert.NoError(t, err)
				assert.Len(t, k.DeleteKeygenSessionCalls(), 1)
				assert.Len(t, pool.ClearRewardsCalls(), 10)
			}).
			Run(t)

		givenKeepersAndCtx.
			When("a completed keygen session expiry equal to the block height", func() {
				k.GetKeygenSessionsByExpiryFunc = func(_ sdk.Context, expiry int64) []types.KeygenSession {
					if expiry != ctx.BlockHeight() {
						return nil
					}

					return []types.KeygenSession{{
						Key:   typestestutils.Key(),
						State: exported.Completed,
					}}
				}
			}).
			Then("should delete and set key", func(t *testing.T) {
				k.DeleteKeygenSessionFunc = func(sdk.Context, exported.KeyID) {}
				k.SetKeyFunc = func(sdk.Context, types.Key) {}

				_, err := multisig.EndBlocker(ctx, abci.RequestEndBlock{}, k, rewarder)

				assert.NoError(t, err)
				assert.Len(t, k.DeleteKeygenSessionCalls(), 1)
				assert.Len(t, k.SetKeyCalls(), 1)
			}).
			Run(t)
	})
}
