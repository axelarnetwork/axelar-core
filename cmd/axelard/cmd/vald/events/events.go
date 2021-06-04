package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	"github.com/cosmos/cosmos-sdk/types"
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
func Consume(subscriber events.FilteredSubscriber, process func(blockHeight int64, attributes []types.Attribute) error) jobs.Job {
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
	state           SyncState
	updateAvailable chan struct{}
}

// NewMgr returns a new mgr instance
func NewMgr(client rpcclient.SignClient, store StateStore, pubsubFactory func() pubsub.Bus, logger log.Logger) *Mgr {
	state := store.Read()
	mgr := &Mgr{
		subscribeLock: sync.RWMutex{},
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
	}

	if state.SeenBlock > state.CompletedBlock {
		mgr.updateAvailable <- struct{}{}
	}
	return mgr
}

func (m *Mgr) FetchEvents(done chan struct{}) <-chan error {
	errChan := make(chan error, 2)
	go func() {
		defer func() { done <- struct{}{} }()
		defer m.logger.Info("shutting down")

		errChan <- m.processUpdates(done)

		m.logger.Info("closing all subscriptions")
		m.subscribeLock.Lock()
		defer m.subscribeLock.Unlock()
		for _, sub := range m.subscriptions {
			sub.Close()
		}

		m.logger.Info("flushing event sync state")
		errChan <- m.store.Persist(m.state)
	}()

	return errChan
}

func (m *Mgr) processUpdates(done chan struct{}) error {
	for range m.updateAvailable {
		select {
		case <-done:
			return nil
		default:
			currBlock := m.state.CompletedBlock + 1
			block, err := m.queryBlockResults(currBlock)
			if err != nil {
				return err
			}
			err = m.publishEvents(block)
			if err != nil {
				return err
			}

			m.state.CompletedBlock++
		}

		m.checkForUpdate()
	}
	return nil
}

func (m *Mgr) checkForUpdate() {
	// no need to lock here: the exact value of SeenBlock doesn't matter and it can only increase monotonically.
	// So even if another goroutine changes the value this check can never go from "update" to "no update"

	if m.state.SeenBlock > m.state.CompletedBlock {
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
	subscription, ok := m.subscriptions[q.String()]
	if !ok {
		m.subscribeLock.Lock()
		defer m.subscribeLock.Unlock()

		bus := m.createBus()
		subscription = struct {
			tmpubsub.Query
			pubsub.Bus
		}{Query: q, Bus: bus}
		m.subscriptions[q.String()] = subscription
	}

	return subscription.Subscribe()
}

func (m *Mgr) NotifyNewBlock(height int64) {
	// it is important to lock here, otherwise in two (or more) concurrent calls the smaller value might win the data race
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	if height > m.state.SeenBlock {
		m.logger.Debug(fmt.Sprintf("block %d added to queue", height))
		m.state.SeenBlock = height
		m.checkForUpdate()
	}
}

// queryBlockResults retrieves the block of given height from tendermint and extracts all tx events
func (m *Mgr) queryBlockResults(height int64) (*coretypes.ResultBlockResults, error) {
	res, err := m.client.BlockResults(context.Background(), &height)
	if err != nil {
		return nil, err
	}
	m.logger.Debug(fmt.Sprintf("received block %d", height))

	return res, nil
}

func (m *Mgr) publishEvents(block *coretypes.ResultBlockResults) error {
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

type SyncState struct {
	CompletedBlock int64
	SeenBlock      int64
}

type StateStore struct {
	rwc io.ReadWriteCloser
}

func NewStateStore(rwc io.ReadWriteCloser) StateStore {
	return StateStore{rwc: rwc}
}

func (s StateStore) Read() SyncState {
	bz, err := io.ReadAll(s.rwc)
	if err != nil {
		return SyncState{}
	}
	var state SyncState
	err = json.Unmarshal(bz, &state)
	if err != nil {
		return SyncState{}
	}
	return state
}

func (s StateStore) Persist(state SyncState) error {
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = s.rwc.Write(bz)
	if err != nil {
		_ = s.rwc.Close()
		return err
	}

	return s.rwc.Close()
}
