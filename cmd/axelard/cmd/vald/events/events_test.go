package events

import (
	"context"
	"encoding/json"
	"io"
	mathRand "math/rand"
	"testing"
	"time"

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

func TestMgr_FetchEvents(t *testing.T) {
	var (
		rwc *mock.ReadWriteSeekTruncateCloserMock
		mgr *Mgr
	)
	setup := func(initialComplete int64) {
		bus := func() pubsub.Bus { return &mock.BusMock{} }
		copied := 0
		rwc = &mock.ReadWriteSeekTruncateCloserMock{
			ReadFunc: func(bz []byte) (int, error) {
				x, err := json.Marshal(initialComplete)
				if err != nil {
					return 0, err
				}

				if copied < len(x) {
					n := copy(bz, x[copied:])
					copied += n
					return n, nil
				}
				return 0, io.EOF
			},
			WriteFunc:    func(bz []byte) (int, error) { return 0, nil },
			CloseFunc:    func() error { return nil },
			SeekFunc:     func(int64, int) (int64, error) { return 0, nil },
			TruncateFunc: func(int64) error { return nil },
		}
		client := &mock.SignClientMock{
			BlockResultsFunc: func(_ context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
				return &coretypes.ResultBlockResults{Height: *height}, nil
			},
		}
		mgr = NewMgr(client, NewStateStore(rwc), bus, log.TestingLogger())
	}

	repeats := 20
	t.Run("stops when done", testutils.Func(func(t *testing.T) {
		setup(0)

		errChan := mgr.FetchEvents()

		mgr.Shutdown()
		for err := range errChan {
			assert.Nil(t, err)
		}
		assert.Len(t, rwc.WriteCalls(), 1)
		assert.Len(t, rwc.CloseCalls(), 1)
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

		var actualCompleted int64
		assert.NoError(t, json.Unmarshal(rwc.WriteCalls()[0].P, &actualCompleted))
		assert.Equal(t, initialCompleted, actualCompleted)
	}).Repeat(repeats))

	t.Run("fetch all available blocks", testutils.Func(func(t *testing.T) {
		initialCompleted := rand.I64Between(0, 10000)
		setup(initialCompleted)

		errChan := mgr.FetchEvents()
		seen := rand.I64Between(initialCompleted+1, initialCompleted+30)
		mgr.NotifyNewBlock(seen)

		// delay so mgr has time to fetch the block
		time.Sleep(1 * time.Millisecond)
		mgr.Shutdown()
		for err := range errChan {
			assert.Nil(t, err)
		}
		var actualCompleted int64
		assert.NoError(t, json.Unmarshal(rwc.WriteCalls()[0].P, &actualCompleted))
		assert.Equal(t, seen, actualCompleted)
	}).Repeat(repeats))
}

func TestMgr_Subscribe(t *testing.T) {
	var (
		mgr            *Mgr
		client         *mock.SignClientMock
		expectedEvents []abci.Event
		rwc            *mock.ReadWriteSeekTruncateCloserMock
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
					expectedEvents = append(expectedEvents, tx.Events...)
				}
				return result, nil
			},
		}
		copied := 0
		rwc = &mock.ReadWriteSeekTruncateCloserMock{
			ReadFunc: func(bz []byte) (int, error) {
				x, err := json.Marshal(initialComplete)
				if err != nil {
					return 0, err
				}

				if copied < len(x) {
					n := copy(bz, x[copied:])
					copied += n
					return n, nil
				}
				return 0, io.EOF
			},
			WriteFunc:    func(bz []byte) (int, error) { return 0, nil },
			CloseFunc:    func() error { return nil },
			SeekFunc:     func(int64, int) (int64, error) { return 0, nil },
			TruncateFunc: func(int64) error { return nil },
		}

		actualEvents := make(chan pubsub.Event, 100000)
		mgr = NewMgr(client, NewStateStore(rwc), func() pubsub.Bus {
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
		time.Sleep(1 * time.Millisecond)
		// closes channels so we can test deterministically
		mgr.Shutdown()

		var actualEvents []abci.Event
		for e := range sub.Events() {
			actualEvents = append(actualEvents, e.(abci.Event))
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
		expectedResult := make(chan []*abci.ResponseDeliverTx, 1)
		client.BlockResultsFunc = func(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
			res, err := mockedResults(ctx, height)
			assert.NoError(t, err)
			expectedResult <- res.TxsResults
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
		var expectedEventsFiltered []abci.Event
		txs := <-expectedResult
		for _, tx := range txs {
			filteredCount++
			if filteredCount%n == 0 {
				expectedEventsFiltered = append(expectedEventsFiltered, tx.Events...)
			}
		}

		// closes channels so we can test deterministically
		mgr.Shutdown()

		var actualEvents []abci.Event
		for e := range sub.Events() {
			actualEvents = append(actualEvents, e.(abci.Event))
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

		mgr.FetchEvents()
		mgr.NotifyNewBlock(completed + 1)

		// delay so mgr has time to fetch the block
		time.Sleep(1 * time.Millisecond)
		// closes channels so we can test deterministically
		mgr.Shutdown()

		var actualEvents []abci.Event
		for e := range sub.Events() {
			actualEvents = append(actualEvents, e.(abci.Event))
		}

		assert.Equal(t, expectedEvents, actualEvents)
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
