package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/jobs"
)

// Consume processes all events from the given subscriber with the given function.
// Do not consume the same subscriber multiple times.
func Consume(subscriber events.FilteredSubscriber, process func(blockHeight int64, attributes []sdk.Attribute) error) jobs.Job {
	return func(errChan chan<- error) {
	loop:
		for {
			select {
			case e := <-subscriber.Events():
				go func() {
					defer recovery(errChan)
					if err := process(e.Height, e.Attributes); err != nil {
						errChan <- err
					}
				}()
			case <-subscriber.Done():
				break loop
			}
		}
	}
}

func recovery(errChan chan<- error) {
	if r := recover(); r != nil {
		errChan <- fmt.Errorf("job panicked:%s", r)
	}
}

// OnlyBlockHeight wraps a function that only depends on block height and makes it compatible with the Consume function
func OnlyBlockHeight(f func(int64)) func(int64, []sdk.Attribute) error {
	return func(h int64, _ []sdk.Attribute) error { f(h); return nil }
}

// OnlyAttributes wraps a function that only depends on event attributes and makes it compatible with the Consume function
func OnlyAttributes(f func([]sdk.Attribute) error) func(int64, []sdk.Attribute) error {
	return func(_ int64, a []sdk.Attribute) error { return f(a) }
}

// Mgr represents an object that receives blocks from a tendermint server and manages queries for events in those blocks
type Mgr struct {
	subscribeLock sync.RWMutex
	stateLock     sync.RWMutex

	store StateStore

	subscriptions map[string]struct {
		tmpubsub.Query
		pubsub.Bus
	}
	client          rpcclient.SignClient
	createBus       func() pubsub.Bus
	logger          log.Logger
	state           syncState
	updateAvailable chan struct{}
	startCleanup    chan struct{}
	cleanupComplete chan struct{}
}

type syncState struct {
	Completed int64
	Seen      int64
}

// NewMgr returns a new mgr instance
func NewMgr(client rpcclient.SignClient, store StateStore, pubsubFactory func() pubsub.Bus, logger log.Logger) *Mgr {
	state := syncState{
		Completed: store.Read(),
		Seen:      0,
	}

	// sanitize input
	if state.Completed < 0 {
		state.Completed = 0
	}

	mgr := &Mgr{
		subscribeLock: sync.RWMutex{},
		stateLock:     sync.RWMutex{},
		client:        client,
		subscriptions: make(map[string]struct {
			tmpubsub.Query
			pubsub.Bus
		}),
		createBus:       pubsubFactory,
		state:           state,
		store:           store,
		logger:          logger.With("listener", "events"),
		updateAvailable: make(chan struct{}, 1),
		startCleanup:    make(chan struct{}),
		cleanupComplete: make(chan struct{}),
	}

	return mgr
}

// FetchEvents asynchronously queries the blockchain for new blocks and publishes all txs events in those blocks to the event manager's subscribers.
// Any occurring events are pushed into the returned error channel.
func (m *Mgr) FetchEvents() <-chan error {
	errChan := make(chan error, 2)
	go func() {
		defer close(errChan)
		defer m.logger.Info("shutting down")

		errChan <- m.processUpdates()

		m.subscribeLock.Lock()
		defer m.subscribeLock.Unlock()

		m.logger.Info("closing all subscriptions")
		for _, sub := range m.subscriptions {
			sub.Close()
		}

		m.logger.Info("flushing event sync state")
		errChan <- m.store.Write(m.state.Completed)
		m.cleanupComplete <- struct{}{}
	}()

	return errChan
}

func (m *Mgr) processUpdates() error {
	for {
		select {
		case <-m.updateAvailable:
			currBlock := m.state.Completed + 1
			block, err := m.queryBlockResults(currBlock)
			if err != nil {
				return err
			}
			err = m.publishEvents(block)
			if err != nil {
				return err
			}

			m.state.Completed++
		case <-m.startCleanup:
			return nil
		}

		m.checkForUpdate()
	}
}

func (m *Mgr) checkForUpdate() {
	// no need to lock here: the exact value of Seen doesn't matter and it can only increase monotonically.
	// So even if another goroutine changes the value this check can never go from "update" to "no update"

	if m.state.Seen > m.state.Completed {
		// multiple places can call this function, so the select statement prevents stalling when updateAvailable is already set
		select {
		case m.updateAvailable <- struct{}{}:
			return
		default:
			return
		}
	}
}

// Subscribe returns an event subscription based on the given query
func (m *Mgr) Subscribe(q tmpubsub.Query) (pubsub.Subscriber, error) {
	// map cannot deal with concurrent read/writes so we lock for the whole function.
	// Alternatively we would have to acquire a read lock first and then replace it with a write lock if the value doesn't exist.
	// We chose the simpler solution here.
	m.subscribeLock.Lock()
	defer m.subscribeLock.Unlock()

	subscription, ok := m.subscriptions[q.String()]
	if !ok {
		subscription = struct {
			tmpubsub.Query
			pubsub.Bus
		}{Query: q, Bus: m.createBus()}
		m.subscriptions[q.String()] = subscription
	}

	return subscription.Subscribe()
}

// NotifyNewBlock notifies the manager that a new block at the given height is available on the blockchain
func (m *Mgr) NotifyNewBlock(height int64) {
	// it is important to lock here, otherwise for two (or more) concurrent calls the smaller value might win the data race
	// and we miss the update trigger
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	if height > m.state.Seen {
		m.logger.Debug(fmt.Sprintf("block %d added to queue", height))
		m.state.Seen = height
		m.checkForUpdate()
	}
}

// Shutdown cleans up the manager's resources. Blocks until it's done.
func (m *Mgr) Shutdown() {
	m.startCleanup <- struct{}{}
	<-m.cleanupComplete
}

func (m *Mgr) queryBlockResults(height int64) (*coretypes.ResultBlockResults, error) {
	res, err := m.client.BlockResults(context.Background(), &height)
	if err != nil {
		return nil, err
	}
	m.logger.Debug(fmt.Sprintf("received block %d", height))

	return res, nil
}

func (m *Mgr) publishEvents(block *coretypes.ResultBlockResults) error {
	// publishing events and adding subscriptions are mutually exclusive operations.
	// This guarantees that a subscription sees all block events or none
	m.subscribeLock.RLock()
	defer m.subscribeLock.RUnlock()

	for _, txRes := range block.TxsResults {
		eventMap := mapifyEvents(txRes.Events)
		for _, subscription := range m.subscriptions {
			match, err := subscription.Query.Matches(eventMap)
			if err != nil {
				return fmt.Errorf("failed to match against query %s: %w", subscription.Query.String(), err)
			}

			if !match {
				continue
			}

			for _, event := range txRes.Events {
				err := subscription.Publish(event)
				if err != nil {
					return err
				}
			}
		}
	}
	m.logger.Debug(fmt.Sprintf("published all tx events for block %d", block.Height))
	return nil
}

func mapifyEvents(events []abci.Event) map[string][]string {
	result := map[string][]string{tm.EventTypeKey: {tm.EventTx}}
	for _, event := range events {
		if len(event.Type) == 0 {
			return nil
		}

		for _, attr := range event.Attributes {
			if len(attr.Key) == 0 {
				continue
			}

			compositeTag := fmt.Sprintf("%s.%s", event.Type, string(attr.Key))
			result[compositeTag] = append(result[compositeTag], string(attr.Value))
		}
	}
	return result
}

// ReadWriteSeekTruncateCloser effectively provides an interface for os.File so the event manager can be unit tested more easily
type ReadWriteSeekTruncateCloser interface {
	io.ReadWriteSeeker
	Truncate(size int64) error
	Close() error
}

// StateStore manages event state persistence
type StateStore struct {
	rw ReadWriteSeekTruncateCloser
}

// NewStateStore returns a new StateStore instance
func NewStateStore(rw ReadWriteSeekTruncateCloser) StateStore {
	return StateStore{rw: rw}
}

// Read returns the block height for which all events have been published
func (s StateStore) Read() (completed int64) {
	bz, err := io.ReadAll(s.rw)
	if err != nil {
		return 0
	}

	err = json.Unmarshal(bz, &completed)
	if err != nil {
		return 0
	}

	return completed
}

// Write persists the block height for which all events have been published
func (s StateStore) Write(completed int64) error {
	bz, err := json.Marshal(completed)
	if err != nil {
		return err
	}

	// overwrite previous value
	_, err = s.rw.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	err = s.rw.Truncate(0)
	if err != nil {
		return err
	}

	_, err = s.rw.Write(bz)
	if err != nil {
		_ = s.rw.Close()
		return err
	}

	return s.rw.Close()
}
