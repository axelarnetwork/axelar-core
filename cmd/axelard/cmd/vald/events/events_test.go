package events

import (
	"context"
	"encoding/json"
	"fmt"
	mathRand "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	tmTypes "github.com/axelarnetwork/tm-events/pkg/tendermint/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestMgr_NewMgr(t *testing.T) {
	bus := func() pubsub.Bus { return &mock.BusMock{} }

	repeats := 20
	t.Run("without state should start from latest block when it is available", testutils.Func(func(t *testing.T) {
		blockHeight := rand.PosI64()

		rw := &mock.ReadWriterMock{ReadAllFunc: func() ([]byte, error) { return nil, fmt.Errorf("some error") }}
		client := &mock.SignClientMock{
			BlockFunc: func(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
				return &coretypes.ResultBlock{Block: &tm.Block{Header: tm.Header{Height: blockHeight}}}, nil
			},
		}

		mgr := NewMgr(client, rw, bus, log.TestingLogger())

		assert.Equal(t, blockHeight, mgr.state.completed)
	}).Repeat(repeats))

	t.Run("without state should start from block 0 when latest block is not available", testutils.Func(func(t *testing.T) {
		rw := &mock.ReadWriterMock{ReadAllFunc: func() ([]byte, error) { return nil, fmt.Errorf("some error") }}

		client := &mock.SignClientMock{
			BlockFunc: func(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
				return &coretypes.ResultBlock{Block: nil}, nil
			},
		}
		mgr := NewMgr(client, rw, bus, log.TestingLogger())
		assert.Equal(t, int64(0), mgr.state.completed)

		client = &mock.SignClientMock{
			BlockFunc: func(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
				return nil, fmt.Errorf("some error")
			},
		}
		mgr = NewMgr(client, rw, bus, log.TestingLogger())
		assert.Equal(t, int64(0), mgr.state.completed)
	}).Repeat(repeats))

	t.Run("should start a block that is persisted", testutils.Func(func(t *testing.T) {
		blockHeight := rand.PosI64()

		rw := &mock.ReadWriterMock{ReadAllFunc: func() ([]byte, error) { return json.Marshal(blockHeight) }}
		client := &mock.SignClientMock{}
		mgr := NewMgr(client, rw, bus, log.TestingLogger())
		assert.Equal(t, blockHeight, mgr.state.completed)
	}).Repeat(repeats))
}

func TestMgr_FetchEvents(t *testing.T) {
	var (
		rw  *mock.ReadWriterMock
		mgr *Mgr
	)
	setup := func(initialComplete int64) {
		bus := func() pubsub.Bus { return &mock.BusMock{} }
		rw = &mock.ReadWriterMock{
			ReadAllFunc:  func() ([]byte, error) { return json.Marshal(initialComplete) },
			WriteAllFunc: func([]byte) error { return nil },
		}
		client := &mock.SignClientMock{
			BlockResultsFunc: func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
				return &coretypes.ResultBlockResults{Height: *height}, nil
			},
		}
		mgr = NewMgr(client, rw, bus, log.TestingLogger())
	}

	repeats := 20
	t.Run("stops when done", testutils.Func(func(t *testing.T) {
		setup(0)

		errChan := mgr.FetchEvents()

		mgr.Shutdown()
		for err := range errChan {
			assert.Nil(t, err)
		}
	}).Repeat(repeats))

	t.Run("do not fetch blocks when no update available", testutils.Func(func(t *testing.T) {
		initialCompleted := rand.PosI64()
		setup(initialCompleted)

		errChan := mgr.FetchEvents()
		mgr.NotifyNewBlock(rand.I64Between(0, initialCompleted+1))

		mgr.Shutdown()
		for err := range errChan {
			assert.Nil(t, err)
		}

		assert.Len(t, rw.WriteAllCalls(), 0)
	}).Repeat(repeats))

	t.Run("fetch all available blocks", testutils.Func(func(t *testing.T) {
		initialCompleted := rand.I64Between(0, 10000)
		setup(initialCompleted)

		errChan := mgr.FetchEvents()
		seen := rand.I64Between(initialCompleted+1, initialCompleted+30)
		done := make(chan struct{})
		rw.WriteAllFunc = func(bz []byte) error {
			var written int64
			err := json.Unmarshal(bz, &written)
			assert.NoError(t, err)
			if written == seen {
				done <- struct{}{}
			}
			return nil
		}
		mgr.NotifyNewBlock(seen)

		assert.NoError(t, waitFor(done))
		mgr.Shutdown()
		for err := range errChan {
			assert.Nil(t, err)
		}
		var actualCompleted int64
		assert.NoError(t, json.Unmarshal(rw.WriteAllCalls()[len(rw.WriteAllCalls())-1].Bytes, &actualCompleted))
		assert.Equal(t, seen, actualCompleted)
	}).Repeat(1))
}

func TestMgr_Subscribe(t *testing.T) {
	var (
		mgr            *Mgr
		client         *mock.SignClientMock
		expectedEvents []tmTypes.Event
		rw             *mock.ReadWriterMock
		query          *mock.QueryMock
	)

	setup := func(initialComplete int64) {
		client = &mock.SignClientMock{
			BlockResultsFunc: func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
				result := &coretypes.ResultBlockResults{
					Height:     *height,
					TxsResults: randomTxResults(rand.I64Between(0, 100)),
				}

				expectedEvents = nil
				for _, tx := range result.TxsResults {
					for _, event := range tx.Events {
						e, ok := events.ProcessEvent(event)
						e.Height = result.Height
						assert.True(t, ok)
						expectedEvents = append(expectedEvents, e)
					}
				}
				return result, nil
			},
		}
		rw = &mock.ReadWriterMock{
			ReadAllFunc:  func() ([]byte, error) { return json.Marshal(initialComplete) },
			WriteAllFunc: func([]byte) error { return nil },
		}

		actualEvents := make(chan pubsub.Event, 100000)
		mgr = NewMgr(client, rw, func() pubsub.Bus {
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
		}, log.TestingLogger())

		query = &mock.QueryMock{
			MatchesFunc: func(map[string][]string) (bool, error) { return true, nil },
			StringFunc:  func() string { return rand.StrBetween(1, 100) },
		}
	}

	repeats := 20
	t.Run("query block with txs", testutils.Func(func(t *testing.T) {
		completed := rand.I64Between(0, 10000)
		setup(completed)

		sub, err := mgr.Subscribe(query)
		assert.NoError(t, err)

		mgr.FetchEvents()
		mgr.NotifyNewBlock(completed + 1)

		// delay so mgr has time to fetch the block
		time.Sleep(2 * time.Millisecond)
		// closes channels so we can test deterministically
		mgr.Shutdown()

		var actualEvents []tmTypes.Event
		for e := range sub.Events() {
			actualEvents = append(actualEvents, e.(tmTypes.Event))
		}

		assert.Equal(t, expectedEvents, actualEvents)
	}).Repeat(repeats))

	t.Run("query block without txs", testutils.Func(func(t *testing.T) {
		completed := rand.I64Between(0, 10000)
		setup(completed)
		client.BlockResultsFunc = func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			return &coretypes.ResultBlockResults{
				Height:     *height,
				TxsResults: nil,
			}, nil
		}

		sub, err := mgr.Subscribe(query)
		assert.NoError(t, err)

		mgr.FetchEvents()
		mgr.NotifyNewBlock(completed + 1)

		// closes channels so we can test deterministically
		mgr.Shutdown()

		var actualEvents []abci.Event
		for e := range sub.Events() {
			actualEvents = append(actualEvents, e.(abci.Event))
		}
		assert.Len(t, actualEvents, 0)
	}).Repeat(repeats))

	t.Run("match only some events", testutils.Func(func(t *testing.T) {
		completed := rand.I64Between(0, 10000)
		setup(completed)
		mockedResults := client.BlockResultsFunc
		expectedResult := make(chan *coretypes.ResultBlockResults, 1)
		client.BlockResultsFunc = func(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			res, err := mockedResults(ctx, height)
			assert.NoError(t, err)
			expectedResult <- res
			return res, err
		}

		eventCount := 0
		n := int(rand.I64Between(1, 10))
		sub, err := mgr.Subscribe(&mock.QueryMock{
			MatchesFunc: func(map[string][]string) (bool, error) {
				eventCount++
				return eventCount%n == 0, nil
			},
			StringFunc: func() string { return rand.StrBetween(1, 100) },
		})
		assert.NoError(t, err)

		mgr.FetchEvents()
		mgr.NotifyNewBlock(completed + 1)

		filteredCount := 0
		var expectedEventsFiltered []tmTypes.Event
		res := <-expectedResult
		for _, tx := range res.TxsResults {
			filteredCount++
			if filteredCount%n == 0 {
				for _, event := range tx.Events {
					e, ok := events.ProcessEvent(event)
					assert.True(t, ok)
					e.Height = res.Height
					expectedEventsFiltered = append(expectedEventsFiltered, e)
				}
			}
		}

		// closes channels so we can test deterministically
		mgr.Shutdown()

		var actualEvents []tmTypes.Event
		for e := range sub.Events() {
			actualEvents = append(actualEvents, e.(tmTypes.Event))
		}
		assert.Equal(t, expectedEventsFiltered, actualEvents)
	}).Repeat(repeats))

	t.Run("match tm.event=Tx", testutils.Func(func(t *testing.T) {
		completed := rand.I64Between(0, 10000)
		setup(completed)

		var err error
		sub, err := mgr.Subscribe(&mock.QueryMock{
			MatchesFunc: func(events map[string][]string) (bool, error) {
				for key, values := range events {
					if key != tm.EventTypeKey {
						continue
					}

					for _, value := range values {
						if value == tm.EventTx {
							return true, nil
						}
					}
				}
				return false, nil
			},
			StringFunc: func() string { return rand.StrBetween(1, 100) },
		})
		assert.NoError(t, err)

		errChan := mgr.FetchEvents()
		mgr.NotifyNewBlock(completed + 1)

		once := sync.Once{}
		var actualEvents []tmTypes.Event
		for e := range sub.Events() {
			// closes once we get events, because then we can be sure the whole block has been processed
			once.Do(mgr.Shutdown)

			actualEvents = append(actualEvents, e.(tmTypes.Event))
		}

		for err := range errChan {
			assert.NoError(t, err)
		}

		assert.Equal(t, expectedEvents, actualEvents)
	}).Repeat(repeats))
}

func TestStateStore_GetState(t *testing.T) {
	repeats := 20
	rw := &mock.ReadWriterMock{}
	store := NewStateStore(rw)

	t.Run("return positive block height", testutils.Func(func(t *testing.T) {
		expected := rand.PosI64()
		rw.ReadAllFunc = func() ([]byte, error) { return json.Marshal(expected) }
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("return block 0", testutils.Func(func(t *testing.T) {
		expected := int64(0)
		rw.ReadAllFunc = func() ([]byte, error) { return json.Marshal(expected) }
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("negative value", testutils.Func(func(t *testing.T) {
		expected := -1 * rand.PosI64()
		rw.ReadAllFunc = func() ([]byte, error) { return json.Marshal(expected) }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("reader error", testutils.Func(func(t *testing.T) {
		rw.ReadAllFunc = func() ([]byte, error) { return nil, fmt.Errorf("some error") }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("wrong data format", testutils.Func(func(t *testing.T) {
		rw.ReadAllFunc = func() ([]byte, error) { return rand.BytesBetween(1, 100), nil }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("empty reader", testutils.Func(func(t *testing.T) {
		rw.ReadAllFunc = func() ([]byte, error) { return []byte{}, nil }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestStateStore_SetState(t *testing.T) {
	repeats := 20
	rw := &mock.ReadWriterMock{}
	store := NewStateStore(rw)

	t.Run("persist positive block height", testutils.Func(func(t *testing.T) {
		var storage []byte
		rw.ReadAllFunc = func() ([]byte, error) { return storage, nil }
		rw.WriteAllFunc = func(bz []byte) error { storage = bz; return nil }
		expected := rand.PosI64()
		assert.NoError(t, store.SetState(expected))
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("persist block 0", testutils.Func(func(t *testing.T) {
		var storage []byte
		rw.ReadAllFunc = func() ([]byte, error) { return storage, nil }
		rw.WriteAllFunc = func(bz []byte) error { storage = bz; return nil }
		expected := int64(0)
		assert.NoError(t, store.SetState(expected))
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("negative value", testutils.Func(func(t *testing.T) {
		rw.WriteAllFunc = func(bz []byte) error { return nil }
		assert.Error(t, store.SetState(-1*rand.PosI64()))
	}).Repeat(repeats))

	t.Run("write fails", testutils.Func(func(t *testing.T) {
		rw.WriteAllFunc = func(bz []byte) error { return fmt.Errorf("some error") }
		assert.Error(t, store.SetState(rand.PosI64()))
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

func waitFor(done <-chan struct{}) error {
	timeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	select {
	case <-done:
		return nil
	case <-timeout.Done():
		return fmt.Errorf("timeout")
	}
}
