package nexus_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	exportedmock "github.com/axelarnetwork/axelar-core/x/nexus/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	rewardexported "github.com/axelarnetwork/axelar-core/x/reward/exported"
	rewardmock "github.com/axelarnetwork/axelar-core/x/reward/exported/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestRouteQueuedMessages(t *testing.T) {
	var (
		msgs     []exported.GeneralMessage
		ctx      sdk.Context
		n        *mock.NexusMock
		reward   *mock.RewardKeeperMock
		snapshot *mock.SnapshotterMock
	)

	givenTheEndBlocker := Given("everything needed for the end blocker", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))

		n = &mock.NexusMock{
			LoggerFunc:    func(_ sdk.Context) log.Logger { return log.NewTestLogger(t) },
			GetChainsFunc: func(ctx sdk.Context) []exported.Chain { return nil },
			GetParamsFunc: func(ctx sdk.Context) types.Params { return types.DefaultParams() },
		}
		reward = &mock.RewardKeeperMock{}
		snapshot = &mock.SnapshotterMock{}
	})

	givenTheEndBlocker.
		When("no route message is queued", func() {
			n.DequeueRouteMessageFunc = func(ctx sdk.Context) (exported.GeneralMessage, bool) { return exported.GeneralMessage{}, false }
		}).
		Then("no message is routed", func(t *testing.T) {
			_, err := nexus.EndBlocker(ctx, n, reward, snapshot)
			assert.NoError(t, err)
		}).
		Run(t)

	givenTheEndBlocker.
		When("some route message is queued", func() {
			msgs = []exported.GeneralMessage{{ID: rand.NormalizedStr(10)}}
			count := 0

			n.DequeueRouteMessageFunc = func(ctx sdk.Context) (exported.GeneralMessage, bool) {
				defer func() { count++ }()

				if count >= len(msgs) {
					return exported.GeneralMessage{}, false
				}

				return msgs[count], true
			}
		}).
		When("the message is routed unsuccessfully", func() {
			n.RouteMessageFunc = func(_ sdk.Context, _ string, _ ...exported.RoutingContext) error {
				return fmt.Errorf("failed routing")
			}
		}).
		Then("the failure should be ignored", func(t *testing.T) {
			_, err := nexus.EndBlocker(ctx, n, reward, snapshot)
			assert.NoError(t, err)

			assert.Len(t, n.RouteMessageCalls(), len(msgs))
			for i, call := range n.RouteMessageCalls() {
				assert.Equal(t, msgs[i].ID, call.ID)
			}
		}).
		Run(t)

	endBlockerLimit := types.DefaultParams().EndBlockerLimit
	givenTheEndBlocker.
		When("some route messages are queued", func() {
			msgs = make([]exported.GeneralMessage, endBlockerLimit+10)
			for i := uint64(0); i < endBlockerLimit+10; i++ {
				msgs[i] = exported.GeneralMessage{ID: rand.NormalizedStr(10)}
			}
			count := 0

			n.DequeueRouteMessageFunc = func(ctx sdk.Context) (exported.GeneralMessage, bool) {
				defer func() { count++ }()

				if count >= len(msgs) {
					return exported.GeneralMessage{}, false
				}

				return msgs[count], true
			}
		}).
		When("the messages are routed successfully", func() {
			n.RouteMessageFunc = func(_ sdk.Context, _ string, _ ...exported.RoutingContext) error {
				return nil
			}
		}).
		Then("should route up to the maximum messages", func(t *testing.T) {
			_, err := nexus.EndBlocker(ctx, n, reward, snapshot)
			assert.NoError(t, err)

			assert.Len(t, n.RouteMessageCalls(), int(endBlockerLimit))
			for i, call := range n.RouteMessageCalls() {
				assert.Equal(t, msgs[i].ID, call.ID)
			}
		}).
		Run(t)
}

func TestCheckChainMaintainers(t *testing.T) {
	var (
		ctx      sdk.Context
		n        *mock.NexusMock
		reward   *mock.RewardKeeperMock
		snapshot *mock.SnapshotterMock
	)

	givenDeregisterableMaintainer := Given("a chain with a maintainer that must be deregistered", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.NewTestLogger(t))
		chain := exported.Chain{Name: exported.ChainName("ethereum")}
		maintainer := rand.ValAddr()

		maintainerState := &exportedmock.MaintainerStateMock{
			GetAddressFunc:          func() sdk.ValAddress { return maintainer },
			CountMissingVotesFunc:   func(int) uint64 { return 0 },
			CountIncorrectVotesFunc: func(int) uint64 { return 0 },
		}

		n = &mock.NexusMock{
			LoggerFunc:           func(_ sdk.Context) log.Logger { return log.NewTestLogger(t) },
			GetChainsFunc:        func(sdk.Context) []exported.Chain { return []exported.Chain{chain} },
			IsChainActivatedFunc: func(sdk.Context, exported.Chain) bool { return true },
			GetParamsFunc:        func(sdk.Context) types.Params { return types.DefaultParams() },
			GetChainMaintainerStatesFunc: func(sdk.Context, exported.Chain) []exported.MaintainerState {
				return []exported.MaintainerState{maintainerState}
			},
			DequeueRouteMessageFunc: func(sdk.Context) (exported.GeneralMessage, bool) {
				return exported.GeneralMessage{}, false
			},
		}
		reward = &mock.RewardKeeperMock{
			GetPoolFunc: func(sdk.Context, string) rewardexported.RewardPool {
				return &rewardmock.RewardPoolMock{ClearRewardsFunc: func(sdk.ValAddress) {}}
			},
		}
		// no active proxy => the maintainer must be deregistered
		snapshot = &mock.SnapshotterMock{
			GetProxyFunc: func(sdk.Context, sdk.ValAddress) (sdk.AccAddress, bool) { return nil, false },
		}
	})

	givenDeregisterableMaintainer.
		When("removing the maintainer fails", func() {
			n.RemoveChainMaintainerFunc = func(sdk.Context, exported.Chain, sdk.ValAddress) error {
				return fmt.Errorf("remove failed")
			}
		}).
		Then("the end blocker recovers and does not halt", func(t *testing.T) {
			assert.NotPanics(t, func() {
				_, err := nexus.EndBlocker(ctx, n, reward, snapshot)
				assert.NoError(t, err)
			})
			assert.Len(t, n.RemoveChainMaintainerCalls(), 1)
		}).
		Run(t)
}
