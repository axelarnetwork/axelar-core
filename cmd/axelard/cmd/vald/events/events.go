package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	store stateStore

	subscriptions map[string]struct {
		tmpubsub.Query
		pubsub.Bus
	}
	client          rpcclient.SignClient
	createBus       func() pubsub.Bus
	logger          log.Logger
	state           *syncState
	startCleanup    chan struct{}
	cleanupComplete chan struct{}
}

// syncState is the thread-safe representation of the state of event synchronisation
type syncState struct {
	// using int64 instead of uint64 although values are always positive because the tendermint rpc expects int64,
	// so we don't have to convert the type every time
	completed int64
	seen      int64

	updateAvailable chan struct{}
	stateLock       sync.RWMutex
}

func newSyncState(completed int64) *syncState {
	return &syncState{
		completed:       completed,
		seen:            0,
		updateAvailable: make(chan struct{}, 1),
		stateLock:       sync.RWMutex{},
	}
}

// NewBlockAvailable unblocks when an unprocessed block is available
func (s *syncState) NewBlockAvailable() <-chan struct{} {
	return s.updateAvailable
}

// LatestCompleted returns the latest processed block
func (s *syncState) LatestCompleted() int64 {
	return s.completed
}

// IncrComplete increments the counter for processed blocks
func (s *syncState) IncrComplete() {
	atomic.AddInt64(&s.completed, 1)
	s.processUpdate()
}

// UpdateSeen updates the highest block that has been seen. Returns true if the new block is higher than the previous one
func (s *syncState) UpdateSeen(seen int64) bool {
	defer s.processUpdate()

	s.stateLock.Lock()
	// assert: this unlocks before processUpdate is called
	defer s.stateLock.Unlock()
	if s.seen < seen {
		s.seen = seen
		return true
	}
	return false
}

func (s *syncState) processUpdate() {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()
	if s.seen > s.completed {
		// the updateAvailable "flag" might already be set, in that case nothing needs to be done
		select {
		case s.updateAvailable <- struct{}{}:
			return
		default:
			return
		}
	}
}

// NewMgr returns a new mgr instance
func NewMgr(client rpcclient.SignClient, stateSource ReadWriter, pubsubFactory func() pubsub.Bus, logger log.Logger) *Mgr {
	mgr := &Mgr{
		subscribeLock: sync.RWMutex{},
		client:        client,
		subscriptions: make(map[string]struct {
			tmpubsub.Query
			pubsub.Bus
		}),
		createBus:       pubsubFactory,
		store:           newStateStore(stateSource),
		logger:          logger.With("listener", "events"),
		startCleanup:    make(chan struct{}),
		cleanupComplete: make(chan struct{}),
	}

	completed, err := mgr.store.GetState()
	if err != nil {
		completed = mgr.queryCurrBlockHeight()
	}
	mgr.state = newSyncState(completed)
	logger.Debug(fmt.Sprintf("starting vald from block %d", mgr.state.completed))

	return mgr
}

// FetchEvents asynchronously queries the blockchain for new blocks and publishes all txs events in those blocks to the event manager's subscribers.
// Any occurring events are pushed into the returned error channel.
func (m *Mgr) FetchEvents() <-chan error {
	errChan := make(chan error, 1)
	go func() {
		defer m.logger.Info("shutting down")
		defer close(errChan)

		for {
			select {
			case <-m.state.NewBlockAvailable():
				block, err := m.queryBlockResults(m.state.LatestCompleted() + 1)
				if err != nil {
					errChan <- err
					return
				}

				if err = m.publishEvents(block); err != nil {
					errChan <- err
					return
				}

				m.state.IncrComplete()

				if err = m.store.SetState(m.state.LatestCompleted()); err != nil {
					errChan <- err
					return
				}
			case <-m.startCleanup:
				m.logger.Info("closing all subscriptions")

				m.subscribeLock.Lock()
				for _, sub := range m.subscriptions {
					sub.Close()
				}
				m.subscribeLock.Unlock()

				m.cleanupComplete <- struct{}{}
				return
			}
		}
	}()

	return errChan
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
	if m.state.UpdateSeen(height) {
		m.logger.Debug(fmt.Sprintf("block %d added to queue", height))
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

func (m *Mgr) queryCurrBlockHeight() int64 {
	latestBlock, err := m.client.Block(context.Background(), nil)
	if err != nil || latestBlock.Block == nil {
		return 0
	}
	return latestBlock.Block.Height
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
				e, ok := events.ProcessEvent(event)
				if !ok {
					return fmt.Errorf("could not parse event %v", event)
				}
				e.Height = block.Height
				err := subscription.Publish(e)
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

// ReadWriter represents a data source/sink
type ReadWriter interface {
	WriteAll([]byte) error
	ReadAll() ([]byte, error)
}

// stateStore manages event state persistence
type stateStore struct {
	rw ReadWriter
}

// newStateStore returns a new stateStore instance
func newStateStore(rw ReadWriter) stateStore {
	return stateStore{rw: rw}
}

// GetState returns the stored block height for which all events have been published
func (s stateStore) GetState() (completed int64, err error) {
	bz, err := s.rw.ReadAll()
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not read the event state")
	}

	err = json.Unmarshal(bz, &completed)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "state is in unexpected format")
	}

	if completed < 0 {
		return 0, sdkerrors.Wrap(err, "state must be a positive integer")
	}

	return completed, nil
}

// SetState persists the block height for which all events have been published
func (s stateStore) SetState(completed int64) error {
	if completed < 0 {
		return fmt.Errorf("state must be a positive integer")
	}

	bz, err := json.Marshal(completed)
	if err != nil {
		return err
	}
	return s.rw.WriteAll(bz)
}
