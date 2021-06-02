package events

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/jobs"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	"github.com/cosmos/cosmos-sdk/types"
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
