package events

import (
	"fmt"

	"github.com/axelarnetwork/c2d2/pkg/pubsub"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/events"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/cmd/vald/jobs"
)

// FilteredSubscriber filters events of a subscriber according to a predicate
type FilteredSubscriber struct {
	pubsub.Subscriber
	eventChan chan types.Event
	predicate func(event types.Event) bool
}

func newFilteredSubscriber(subscriber pubsub.Subscriber, predicate func(event types.Event) bool) FilteredSubscriber {
	s := FilteredSubscriber{Subscriber: subscriber, predicate: predicate}

	go func() {
		for event := range s.Subscriber.Events() {
			switch e := event.(type) {
			case types.Event:
				if predicate(e) {
					s.eventChan <- e
				}
			default:
				panic(fmt.Sprintf("unexpected event type %t", event))
			}
		}
	}()
	return s
}

// Events returns a channel of filtered events
func (s FilteredSubscriber) Events() <-chan types.Event {
	return s.eventChan
}

// NewJob creates a job that processes incoming events concurrently
func NewJob(subscriber FilteredSubscriber, process func(attributes []sdk.Attribute) error) jobs.Job {
	return func(errChan chan<- error) {
	loop:
		for {
			select {
			case e := <-subscriber.Events():
				go func() {
					if err := process(e.Attributes); err != nil {
						errChan <- err
					}
				}()
			case <-subscriber.Done():
				break loop
			}
		}
	}
}

// MustSubscribe panics if Subscribe fails
func MustSubscribe(hub *events.Hub, eventType string, module string, action string) FilteredSubscriber {
	subscriber, err := Subscribe(hub, eventType, module, action)
	if err != nil {
		panic(sdkerrors.Wrapf(err, "subscription to event {type %s, module %s, action %s} failed", eventType, module, action))
	}
	return subscriber
}

// Subscribe returns a filtered subscriber that only streams events of the given type, module and action
func Subscribe(hub *events.Hub, eventType string, module string, action string) (FilteredSubscriber, error) {
	bus, err := hub.Subscribe(query.MustParse(fmt.Sprintf("%s='%s' AND %s.%s='%s'",
		tm.EventTypeKey, tm.EventTx, eventType, sdk.AttributeKeyModule, module)))
	if err != nil {
		return FilteredSubscriber{}, err
	}
	subscriber, err := bus.Subscribe()
	if err != nil {
		return FilteredSubscriber{}, err
	}
	return newFilteredSubscriber(
		subscriber,
		func(e types.Event) bool { return e.Type == eventType && e.Module == module && e.Action == action },
	), nil
}
