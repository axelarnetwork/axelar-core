package events

import (
	"context"
	"fmt"
	mathRand "math/rand"
	"testing"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestMgr_QueryEvents(t *testing.T) {
	var (
		sub            pubsub.Subscriber
		mgr            *Mgr
		client         *mock.SignClientMock
		expectedEvents []abci.Event
	)

	setup := func() {
		client = &mock.SignClientMock{
			BlockResultsFunc: func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
				result := &coretypes.ResultBlockResults{
					Height:     *height,
					TxsResults: randomTxResults(rand.I64Between(0, 100)),
				}

				expectedEvents = nil
				for _, tx := range result.TxsResults {
					expectedEvents = append(expectedEvents, tx.Events...)
				}
				return result, nil
			},
		}
		actualEvents := make(chan pubsub.Event, 100000)
		mgr = NewMgr(client, func() pubsub.Bus {
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
			}
		}, log.TestingLogger())
	}
	subscribe := func() {
		var err error
		sub, err = mgr.Subscribe(&mock.QueryMock{
			MatchesFunc: func(map[string][]string) (bool, error) { return true, nil },
			StringFunc:  func() string { return rand.StrBetween(1, 100) },
		})
		if err != nil {
			panic(err)
		}
	}
	repeats := 20
	t.Run("query block with txs", testutils.Func(func(t *testing.T) {
		setup()
		subscribe()

		assert.NoError(t, mgr.QueryEvents(rand.PosI64()))
		assert.Len(t, client.BlockResultsCalls(), 1)
		var actualEvents []abci.Event
	loop:
		for {
			select {
			case e := <-sub.Events():
				actualEvents = append(actualEvents, e.(abci.Event))
			default:
				break loop
			}
		}
		assert.Equal(t, expectedEvents, actualEvents)
	}).Repeat(repeats))

	t.Run("query block without txs", testutils.Func(func(t *testing.T) {
		setup()
		subscribe()
		client.BlockResultsFunc = func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			return &coretypes.ResultBlockResults{
				Height:     *height,
				TxsResults: nil,
			}, nil
		}

		assert.NoError(t, mgr.QueryEvents(rand.PosI64()))
		var actualEvents []abci.Event
	loop:
		for {
			select {
			case e := <-sub.Events():
				actualEvents = append(actualEvents, e.(abci.Event))
			default:
				break loop
			}
		}
		assert.Len(t, actualEvents, 0)
	}).Repeat(repeats))

	t.Run("match only some events", testutils.Func(func(t *testing.T) {
		setup()
		mockedResults := client.BlockResultsFunc
		var expectedResult *coretypes.ResultBlockResults
		client.BlockResultsFunc = func(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			res, err := mockedResults(ctx, height)
			expectedResult = res
			return res, err
		}

		eventCount := 0
		n := int(rand.I64Between(1, 10))
		var err error
		sub, err = mgr.Subscribe(&mock.QueryMock{
			MatchesFunc: func(map[string][]string) (bool, error) {
				eventCount++
				return eventCount%n == 0, nil
			},
			StringFunc: func() string { return rand.StrBetween(1, 100) },
		})
		assert.NoError(t, err)
		assert.NoError(t, mgr.QueryEvents(rand.PosI64()))

		filteredCount := 0
		var expectedEventsFiltered []abci.Event
		for _, tx := range expectedResult.TxsResults {
			filteredCount++
			if filteredCount%n == 0 {
				expectedEventsFiltered = append(expectedEventsFiltered, tx.Events...)
			}
		}

		var actualEvents []abci.Event
	loop:
		for {
			select {
			case e := <-sub.Events():
				actualEvents = append(actualEvents, e.(abci.Event))
			default:
				break loop
			}
		}

		assert.Equal(t, expectedEventsFiltered, actualEvents)
	}).Repeat(repeats))

	t.Run("match tm.event=Tx", testutils.Func(func(t *testing.T) {
		setup()

		var err error
		sub, err = mgr.Subscribe(&mock.QueryMock{
			MatchesFunc: func(events map[string][]string) (bool, error) {
				for key, values := range events {
					if key == tm.EventTypeKey {
						for _, value := range values {
							if value == tm.EventTx {
								return true, nil
							}
						}
					}
				}
				return false, nil
			},
			StringFunc: func() string { return rand.StrBetween(1, 100) },
		})
		assert.NoError(t, err)
		assert.NoError(t, mgr.QueryEvents(rand.PosI64()))

		var actualEvents []abci.Event
	loop:
		for {
			select {
			case e := <-sub.Events():
				actualEvents = append(actualEvents, e.(abci.Event))
			default:
				break loop
			}
		}

		assert.Equal(t, expectedEvents, actualEvents)
	}).Repeat(repeats))

	t.Run("no subscriptions", testutils.Func(func(t *testing.T) {
		setup()

		assert.NoError(t, mgr.QueryEvents(rand.PosI64()))
	}).Repeat(repeats))

	t.Run("query block with negative block number", testutils.Func(func(t *testing.T) {
		setup()
		client.BlockResultsFunc = func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			assert.True(t, *height < 0)
			return nil, fmt.Errorf("negative block numbers not allowed")
		}

		assert.Error(t, mgr.QueryEvents(-1*rand.PosI64()))
	}).Repeat(repeats))

	t.Run("query block with future block number", testutils.Func(func(t *testing.T) {
		setup()
		currHeight := rand.I64Between(0, 10e8)
		client.BlockResultsFunc = func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			assert.True(t, *height > currHeight)
			return nil, fmt.Errorf("negative block numbers not allowed")
		}

		assert.Error(t, mgr.QueryEvents(rand.I64Between(currHeight+1, 10e12)))
	}).Repeat(repeats))
}

func randomTxResults(count int64) []*abci.ResponseDeliverTx {
	resp := make([]*abci.ResponseDeliverTx, 0, count)
	for i := 0; i < cap(resp); i++ {
		resp = append(resp, &abci.ResponseDeliverTx{
			Code:      mathRand.Uint32(),
			Data:      rand.Bytes(int(rand.I64Between(100, 200))),
			Log:       rand.StrBetween(5, 100),
			Info:      rand.StrBetween(5, 100),
			GasWanted: rand.PosI64(),
			GasUsed:   rand.PosI64(),
			Events:    randomEvents(rand.I64Between(1, 10)),
			Codespace: rand.StrBetween(5, 100),
		})
	}

	return resp
}

func randomEvents(count int64) []abci.Event {
	events := make([]abci.Event, 0, count)
	for i := 0; i < cap(events); i++ {
		events = append(events, abci.Event{
			Type:       tm.EventTx,
			Attributes: randomAttributes(rand.I64Between(1, 10)),
		})
	}
	return events
}

func randomAttributes(count int64) []abci.EventAttribute {
	attributes := make([]abci.EventAttribute, 0, count)
	for i := 0; i < cap(attributes); i++ {
		attributes = append(attributes, abci.EventAttribute{
			Key:   rand.BytesBetween(5, 100),
			Value: rand.BytesBetween(5, 100),
			Index: rand.Bools(0.5).Next(),
		})
	}
	return attributes
}
