package vald

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	tmEvents "github.com/axelarnetwork/tm-events/events"
)

// proof of concept for the panic mechanism used in the listen(...) function to panic when it takes too long to see new blocks
func TestPanic(t *testing.T) {
	testTimeout, testCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer testCancel()

	assert.Panics(t, func() {
		timer := time.AfterFunc(0, func() {})
		defer timer.Stop()
		blockTimeout, timeoutCancel := context.WithCancel(context.Background())
		var blocksSeen atomic.Uint64 // Atomic type is used to prevent a false positive data race error.
		newBlock := func() {
			timer.Stop()
			timer = time.AfterFunc(1*time.Millisecond, func() {
				timeoutCancel()
			})
			blocksSeen.Add(1)
		}

		go func() {
			for i := 0; i < 100; i++ {
				newBlock()
			}
			time.Sleep(10 * time.Millisecond)
			newBlock()
		}()

		select {
		case <-testTimeout.Done():
			return
		case <-blockTimeout.Done():
			assert.Equal(t, uint64(100), blocksSeen.Load())
			panic("no new blocks discovered, is the chain halted?")
		}
	})
}

// TestEventFiltering is a regression test for the event filter bug where KeygenStarted
// events were incorrectly matching the SigningStarted filter due to proto package incompatibility.
// See: https://github.com/axelarnetwork/tm-events/pull/64
func TestEventFiltering(t *testing.T) {
	app.SetConfig()

	// event from an actual devnet run that made vald panic because the filter let it through to the signing manager
	keygenEvent := abci.Event{
		Type: "axelar.multisig.v1beta1.KeygenStarted",
		Attributes: []abci.EventAttribute{
			{Key: "key_id", Value: `"ganache-0-genesis"`, Index: true},
			{Key: "module", Value: `"multisig"`, Index: true},
			{Key: "participants", Value: `["axelarvaloper19d6df2msak8eh9j4ehrs4je9hq64f3e6jjvnza","axelarvaloper19wcjfyvvw3yhpf65u3t2d0n6ekj7jgsgn9ynzp","axelarvaloper1fcxl0l3mse6eqg3ewgkx7n8l2dpgvaeyglruh8","axelarvaloper10xen9gatvyyt8capt2r9c4sjw46u5ctld5zvj2","axelarvaloper1av825z9wnty9nwu3a0tp73lc46gnng0m04demv"]`, Index: true},
			{Key: "msg_index", Value: "0", Index: true},
		},
	}

	eventWithHeight := tmEvents.ABCIEventWithHeight{
		Event:  keygenEvent,
		Height: 1832,
	}

	// Create filters for both event types
	keygenFilter := tmEvents.Filter[*types.KeygenStarted]()
	signingFilter := tmEvents.Filter[*types.SigningStarted]()

	keygenMatches := keygenFilter(eventWithHeight)
	signingMatches := signingFilter(eventWithHeight)

	// KeygenStarted event should match KeygenStarted filter
	assert.True(t, keygenMatches, "KeygenStarted event should match KeygenStarted filter")

	// KeygenStarted event should NOT match SigningStarted filter
	// This was the bug: when using gogo/protobuf instead of gogoproto, proto.MessageName()
	// returned empty strings, causing all filters to match all events
	assert.False(t, signingMatches, "KeygenStarted event should NOT match SigningStarted filter")
}
