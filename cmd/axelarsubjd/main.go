package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/axelarnetwork/c2d2/pkg/pubsub"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/client"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/events"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	tm "github.com/tendermint/tendermint/types"

	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

func main() {
	conf := client.Config{
		Address:  client.DefaultAddress,
		Endpoint: client.DefaultWSEndpoint,
	}

	c, err := client.NewConnectedClient(conf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	hub := events.NewHub(c)

	fmt.Println("Start listening to events")

	err = listen(&hub)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Shutting down")
}

func listen(hub *events.Hub) error {
	keygen, err := subscribeToEvent(hub, tss.EventTypeKeygen, tss.ModuleName, tss.AttributeValueStart)
	if err != nil {
		return err
	}
	sign, err := subscribeToEvent(hub, tss.EventTypeSign, tss.ModuleName, tss.AttributeValueStart)
	if err != nil {
		return err
	}

	processors := []func(){func() { processKeygen(keygen) }, func() { processSign(sign) }}

	wg := &sync.WaitGroup{}
	wg.Add(len(processors))
	for _, process := range processors {
		go func(f func()) {
			defer wg.Done()
			f()
		}(process)
	}
	wg.Wait()

	return nil
}

func subscribeToEvent(hub *events.Hub, eventType string, module string, action string) (pubsub.Subscriber, error) {
	bus, err := hub.Subscribe(query.MustParse(fmt.Sprintf("%s='%s' AND %s.%s='%s' AND %s.%s='%s'",
		tm.EventTypeKey, tm.EventTx, eventType, sdk.AttributeKeyAction, action, eventType, sdk.AttributeKeyModule, module)))
	if err != nil {
		return nil, err
	}
	subscriber, err := bus.Subscribe()
	if err != nil {
		return nil, err
	}
	return subscriber, nil
}

func processKeygen(k pubsub.Subscriber) {
	for {
		select {
		case event := <-k.Events():
			switch e := event.(type) {
			case types.Event:
				// all events of the transaction are returned, so need to filter for keygen
				if e.Type == tss.EventTypeKeygen {
					// TODO: move tss keygen msg management here
					fmt.Println("Keygen event:")
					fmt.Println(event)
				}
			default:
				panic(fmt.Sprintf("unexpected event type %t", event))
			}
		case <-k.Done():
			break
		}
	}
}

func processSign(s pubsub.Subscriber) {
	for {
		select {
		case event := <-s.Events():
			switch e := event.(type) {
			case types.Event:
				// all events of the transaction are returned, so need to filter for sign
				if e.Type == tss.EventTypeSign {
					// TODO: move tss sign msg management here
					fmt.Println("Sign event:")
					fmt.Println(event)
				}
			default:
				panic(fmt.Sprintf("unexpected event type %t", event))
			}
		case <-s.Done():
			break
		}
	}
}
