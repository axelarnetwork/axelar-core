package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	"github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/jobs"
)

// Consume processes all events from the given subscriber with the given function.
// Do not consume the same subscriber multiple times.
func Consume(subscriber events.FilteredSubscriber, process func(blockHeight int64, attributes []sdk.Attribute) error) jobs.Job {
	return func(errChan chan<- error) {
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
				return
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

// EventBus represents an object that receives blocks from a tendermint server and manages queries for events in those blocks
type EventBus struct {
	subscribeLock sync.RWMutex

	source BlockSource

	subscriptions map[string]struct {
		tmpubsub.Query
		pubsub.Bus
	}
	createBus func() pubsub.Bus
	logger    log.Logger

	running  context.Context
	shutdown context.CancelFunc
	done     chan struct{}
}

// NewEventBus returns a new event bus instance
func NewEventBus(source BlockSource, pubsubFactory func() pubsub.Bus, logger log.Logger) *EventBus {
	mgr := &EventBus{
		subscribeLock: sync.RWMutex{},
		source:        source,
		subscriptions: make(map[string]struct {
			tmpubsub.Query
			pubsub.Bus
		}),
		createBus: pubsubFactory,
		logger:    logger.With("listener", "events"),
		done:      make(chan struct{}),
	}

	return mgr
}

// FetchEvents asynchronously queries the blockchain for new blocks and publishes all txs events in those blocks to the event manager's subscribers.
// Any occurring errors are pushed into the returned error channel.
func (m *EventBus) FetchEvents(ctx context.Context) <-chan error {
	// either the block source or the event manager could push an error at the same time, so we need to make sure we don't block
	errChan := make(chan error, 2)

	m.running, m.shutdown = context.WithCancel(ctx)
	blockResults, blockErrs := m.source.BlockResults(m.running)

	go func() {
		defer m.shutdown()
		select {
		case err := <-blockErrs:
			errChan <- err
			return
		case <-m.running.Done():
			return
		}
	}()

	go func() {
		defer m.shutdown()
		defer m.logger.Info("shutting down")

		for {
			select {
			case block, ok := <-blockResults:
				if !ok {
					return
				}
				if err := m.publishEvents(block); err != nil {
					errChan <- err
					return
				}
			case <-m.running.Done():
				m.logger.Info("closing all subscriptions")

				m.subscribeLock.Lock()
				for _, sub := range m.subscriptions {
					sub.Close()
				}
				m.subscribeLock.Unlock()
				close(m.done)
				return
			}
		}
	}()

	return errChan
}

// Subscribe returns an event subscription based on the given query
func (m *EventBus) Subscribe(q tmpubsub.Query) (pubsub.Subscriber, error) {
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

// Done returns a channel that gets closed when the EventBus is done cleaning up
func (m *EventBus) Done() <-chan struct{} {
	return m.done
}

func (m *EventBus) publishEvents(block *coretypes.ResultBlockResults) error {
	// Publishing events and adding subscriptions are mutually exclusive operations.
	// This guarantees that a subscription sees all block events or none
	m.subscribeLock.RLock()
	defer m.subscribeLock.RUnlock()

	// beginBlock and endBlock events are published together as block events
	blockEvents := append(block.BeginBlockEvents, block.EndBlockEvents...)
	eventMap := mapifyEvents(blockEvents)
	eventMap[tm.EventTypeKey] = append(eventMap[tm.EventTypeKey], tm.EventNewBlockHeader, tm.EventNewBlock)
	err := m.publishMatches(blockEvents, eventMap, block.Height)
	if err != nil {
		return err
	}

	for _, txRes := range block.TxsResults {
		eventMap = mapifyEvents(txRes.Events)
		eventMap[tm.EventTypeKey] = append(eventMap[tm.EventTypeKey], tm.EventTx)
		err := m.publishMatches(txRes.Events, eventMap, block.Height)
		if err != nil {
			return err
		}
	}

	m.logger.Debug(fmt.Sprintf("published all events for block %d", block.Height))
	return nil
}

func (m *EventBus) publishMatches(abciEvents []abci.Event, eventMap map[string][]string, blockHeight int64) error {
	for _, subscription := range m.subscriptions {
		match, err := subscription.Query.Matches(eventMap)
		if err != nil {
			return fmt.Errorf("failed to match against query %s: %w", subscription.Query.String(), err)
		}

		if !match {
			continue
		}

		for _, abciEvent := range abciEvents {
			event, ok := events.ProcessEvent(abciEvent)
			if !ok {
				return fmt.Errorf("could not parse event %v", abciEvent)
			}
			event.Height = blockHeight
			err := subscription.Publish(event)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func mapifyEvents(events []abci.Event) map[string][]string {
	result := make(map[string][]string)
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
