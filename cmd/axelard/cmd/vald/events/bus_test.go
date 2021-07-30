package events_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	tmEvents "github.com/axelarnetwork/tm-events/pkg/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestMgr_FetchEvents(t *testing.T) {

	t.Run("WHEN the event source throws an error THEN the bus returns error", func(t *testing.T) {
		bus := func() pubsub.Bus { return &mock.BusMock{} }
		errors := make(chan error, 1)
		source := &mock.BlockSourceMock{
			BlockResultsFunc: func(ctx context.Context) (<-chan *coretypes.ResultBlockResults, <-chan error) {
				return make(chan *coretypes.ResultBlockResults), errors
			},
		}
		mgr := events.NewEventBus(source, bus, log.TestingLogger())

		errChan := mgr.FetchEvents(context.Background())

		errors <- fmt.Errorf("some error")

		err := <-errChan
		assert.Error(t, err)
	})

	t.Run("WHEN the block source block result channel closes THEN the bus shuts down", func(t *testing.T) {
		busMock := &mock.BusMock{
			SubscribeFunc: func() (pubsub.Subscriber, error) {
				return &mock.SubscriberMock{}, nil
			},
			CloseFunc: func() {},
		}
		busFactory := func() pubsub.Bus { return busMock }
		results := make(chan *coretypes.ResultBlockResults)
		source := &mock.BlockSourceMock{
			BlockResultsFunc: func(ctx context.Context) (<-chan *coretypes.ResultBlockResults, <-chan error) {
				return results, nil
			},
		}
		mgr := events.NewEventBus(source, busFactory, log.TestingLogger())

		mgr.FetchEvents(context.Background())

		close(results)

		timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		select {
		case <-mgr.Done():
			return
		case <-timeout.Done():
			assert.FailNow(t, "timed out")
		}
	})
}

func TestMgr_Subscribe(t *testing.T) {
	var (
		mgr       *events.EventBus
		query     *mock.QueryMock
		newBlocks chan *coretypes.ResultBlockResults
	)

	setup := func() {
		newBlocks = make(chan *coretypes.ResultBlockResults, 10000)
		source := &mock.BlockSourceMock{BlockResultsFunc: func(ctx context.Context) (<-chan *coretypes.ResultBlockResults, <-chan error) {
			return newBlocks, nil
		}}

		actualEvents := make(chan pubsub.Event, 100000)
		busFactory := func() pubsub.Bus {
			return &mock.BusMock{
				PublishFunc: func(event pubsub.Event) error {
					actualEvents <- event
					return nil
				},
				SubscribeFunc: func() (pubsub.Subscriber, error) {
					return &mock.SubscriberMock{
						EventsFunc: func() <-chan pubsub.Event { return actualEvents },
					}, nil
				},
				CloseFunc: func() {
					close(actualEvents)
				},
			}
		}
		mgr = events.NewEventBus(source, busFactory, log.TestingLogger())

		query = &mock.QueryMock{
			MatchesFunc: func(map[string][]string) (bool, error) { return true, nil },
			StringFunc:  func() string { return rand.StrBetween(1, 100) },
		}
	}

	repeats := 20
	t.Run("query block with txs", testutils.Func(func(t *testing.T) {
		setup()

		mgr.FetchEvents(context.Background())
		sub, err := mgr.Subscribe(query)
		assert.NoError(t, err)

		newBlock := &coretypes.ResultBlockResults{
			Height:           rand.PosI64(),
			BeginBlockEvents: randomEvents(rand.I64Between(0, 10)),
			TxsResults:       randomTxResults(rand.I64Between(1, 10)),
			EndBlockEvents:   randomEvents(rand.I64Between(0, 10)),
		}

		endMarkerBlock := &coretypes.ResultBlockResults{
			Height:           0,
			BeginBlockEvents: randomEvents(rand.I64Between(3, 10)),
			TxsResults:       randomTxResults(rand.I64Between(1, 10)),
			EndBlockEvents:   randomEvents(rand.I64Between(3, 10)),
		}
		newBlocks <- newBlock
		newBlocks <- endMarkerBlock

		timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		expectedEventCount := len(newBlock.BeginBlockEvents) + len(newBlock.EndBlockEvents)
		for _, result := range newBlock.TxsResults {
			expectedEventCount += len(result.Events)
		}
		var eventCount int
		for {
			select {
			case <-timeout.Done():
				assert.FailNow(t, "timed out")
			case event := <-sub.Events():
				assert.IsType(t, tmEvents.Event{}, event)
				actualHeight := event.(tmEvents.Event).Height
				if actualHeight == 0 {
					assert.Equal(t, expectedEventCount, eventCount)
					return
				}
				assert.Equal(t, newBlock.Height, actualHeight)
				eventCount++
			}
		}
	}).Repeat(repeats))

	t.Run("match tm.event=Tx", testutils.Func(func(t *testing.T) {
		setup()

		mgr.FetchEvents(context.Background())
		query.MatchesFunc = func(events map[string][]string) (bool, error) {
			types := events[tm.EventTypeKey]

			for _, t := range types {
				if t == tm.EventTx {
					return true, nil
				}
			}
			return false, nil
		}
		sub, err := mgr.Subscribe(query)
		assert.NoError(t, err)

		newBlock := &coretypes.ResultBlockResults{
			Height:           rand.PosI64(),
			BeginBlockEvents: randomEvents(rand.I64Between(0, 10)),
			TxsResults:       randomTxResults(rand.I64Between(1, 10)),
			EndBlockEvents:   randomEvents(rand.I64Between(0, 10)),
		}

		endMarkerBlock := &coretypes.ResultBlockResults{
			Height:           0,
			BeginBlockEvents: randomEvents(rand.I64Between(3, 10)),
			TxsResults:       randomTxResults(rand.I64Between(1, 10)),
			EndBlockEvents:   randomEvents(rand.I64Between(3, 10)),
		}
		newBlocks <- newBlock
		newBlocks <- endMarkerBlock

		timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		expectedEventCount := 0
		for _, result := range newBlock.TxsResults {
			expectedEventCount += len(result.Events)
		}
		var eventCount int
		for {
			select {
			case <-timeout.Done():
				assert.FailNow(t, "timed out")
			case event := <-sub.Events():
				assert.IsType(t, tmEvents.Event{}, event)
				actualHeight := event.(tmEvents.Event).Height
				if actualHeight == 0 {
					assert.Equal(t, expectedEventCount, eventCount)
					return
				}
				assert.Equal(t, newBlock.Height, actualHeight)
				eventCount++
			}
		}
	}).Repeat(repeats))
}
