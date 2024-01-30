package nexus_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
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
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		n = &mock.NexusMock{
			LoggerFunc:    func(_ sdk.Context) log.Logger { return log.TestingLogger() },
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
			_, err := nexus.EndBlocker(ctx, abci.RequestEndBlock{}, n, reward, snapshot)
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
			_, err := nexus.EndBlocker(ctx, abci.RequestEndBlock{}, n, reward, snapshot)
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
			_, err := nexus.EndBlocker(ctx, abci.RequestEndBlock{}, n, reward, snapshot)
			assert.NoError(t, err)

			assert.Len(t, n.RouteMessageCalls(), int(endBlockerLimit))
			for i, call := range n.RouteMessageCalls() {
				assert.Equal(t, msgs[i].ID, call.ID)
			}
		}).
		Run(t)
}
