package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	"github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
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
	subscriptions map[string]struct {
		tmpubsub.Query
		pubsub.Bus
	}
	mtx       sync.RWMutex
	client    rpcclient.SignClient
	createBus func() pubsub.Bus
	logger    log.Logger
}

// NewMgr returns a new mgr instance
func NewMgr(client rpcclient.SignClient, pubsubFactory func() pubsub.Bus, logger log.Logger) *Mgr {
	return &Mgr{
		client: client,
		subscriptions: make(map[string]struct {
			tmpubsub.Query
			pubsub.Bus
		}),
		mtx:       sync.RWMutex{},
		createBus: pubsubFactory,
		logger:    logger.With("listener", "events"),
	}
}

// Subscribe returns an event subscription based on the given query
func (m *Mgr) Subscribe(q tmpubsub.Query) (pubsub.Subscriber, error) {
	subscription, ok := m.subscriptions[q.String()]
	if !ok {
		m.mtx.Lock()
		defer m.mtx.Unlock()

		bus := m.createBus()
		subscription = struct {
			tmpubsub.Query
			pubsub.Bus
		}{Query: q, Bus: bus}
		m.subscriptions[q.String()] = subscription
	}

	return subscription.Subscribe()
}

// QueryTxEvents retrieves the block of given height from tendermint and extracts all tx events
func (m *Mgr) QueryTxEvents(height int64) error {
	res, err := m.client.BlockResults(context.Background(), &height)
	if err != nil {
		return err
	}

	m.logger.Debug(fmt.Sprintf("received block %d", height))
	for _, txRes := range res.TxsResults {
		err := m.publish(txRes.Events)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Mgr) publish(events []abci.Event) error {
	eventMap := mapifyEvents(events)
	for _, subscription := range m.subscriptions {
		match, err := subscription.Query.Matches(eventMap)
		if err != nil {
			return fmt.Errorf("failed to match against query %s: %w", subscription.Query.String(), err)
		}

		if !match {
			continue
		}

		for _, event := range events {
			err := subscription.Publish(event)
			if err != nil {
				return err
			}
		}
	}
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
