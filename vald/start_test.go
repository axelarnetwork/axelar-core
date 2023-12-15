package vald

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
