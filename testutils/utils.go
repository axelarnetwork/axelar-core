// Package testutils provides general purpose utility functions for unit/integration testing.
package testutils

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	btcTypes "github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	stTypes "github.com/axelarnetwork/axelar-core/x/staking/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	axTypes "github.com/axelarnetwork/axelar-core/x/voting/types"
)

var (
	cdc *codec.Codec
)

// Codec creates a codec for testing with all necessary types registered.
// This codec is not sealed so tests can add their own mock types.
func Codec() *codec.Codec {
	// Use cache if initialized before
	if cdc != nil {
		return cdc
	}

	cdc = codec.New()

	sdk.RegisterCodec(cdc)

	// Add new modules here so tests have access to marshalling the registered types
	axTypes.RegisterCodec(cdc)
	btcTypes.RegisterCodec(cdc)
	tssTypes.RegisterCodec(cdc)
	broadcastTypes.RegisterCodec(cdc)
	stTypes.RegisterCodec(cdc)

	return cdc
}

// RandIntGen represents an random integer generator.
// Call Stop when done so dangling goroutines can be cleaned up.
type RandIntGen struct {
	ch      chan int
	done    chan struct{}
	wrapped *RandIntGen
}

// RandInts returns a random integer generator for positive integers.
func RandInts() RandIntGen {
	return generate(rand.Int)
}

// RandIntsBetween returns a random integer generator for numbers between lower (inclusive) and upper (exclusive).
// It panics if  upper <= lower.
func RandIntsBetween(lower int, upper int) RandIntGen {
	return generate(func() int { return rand.Intn(upper-lower) + lower })
}

// Restrict the output of the underlying generator to adhere to the predicate.
// If the predicate is not satisfiable the Take function will deadlock.
func (g RandIntGen) Where(predicate func(i int) bool) RandIntGen {
	newGen := RandIntGen{ch: make(chan int), wrapped: &g}
	go func() {
		// cascade channel close when underlying generator channel closes
		defer close(newGen.ch)
		for n := range g.ch {
			if predicate(n) {
				newGen.ch <- n
			}
		}
	}()
	return newGen
}

// Take returns a slice of random integers of the given length.
func (g RandIntGen) Take(count int) []int {
	nums := make([]int, 0, count)
	for i := 0; i < count; i++ {
		nums = append(nums, <-g.ch)
	}
	return nums
}

// Next returns a single random integer.
func (g RandIntGen) Next() int {
	return <-g.ch
}

// Stop closes all goroutines used during number generation.
func (g *RandIntGen) Stop() {
	// stop the deepest wrapped channel in
	if g.wrapped != nil {
		g.wrapped.Stop()
	} else {
		close(g.done)
	}

	// The underlying generator might be stuck in the default select case trying to push a value into the channel,
	// so we need to make sure it is unstuck to be able to close the output channel
	<-g.ch
}

func generate(generator func() int) RandIntGen {
	g := RandIntGen{ch: make(chan int), done: make(chan struct{}), wrapped: nil}
	go func() {
		for {
			select {
			case <-g.done:
				close(g.ch)
				return
			default:
				g.ch <- generator()
			}
		}
	}()
	return g
}

// RandBoolGen represents an random bool generator.
// Call Stop when done so dangling goroutines can be cleaned up.
type RandBoolGen struct {
	ch   chan bool
	done chan struct{}
}

// RandBools returns a random bool generator that adheres to the given ratio of true to false values.
func RandBools(ratio float64) RandBoolGen {
	g := RandBoolGen{ch: make(chan bool), done: make(chan struct{})}
	go func() {
		for {
			select {
			case <-g.done:
				close(g.ch)
				return
			default:
				g.ch <- rand.Float64() < ratio
			}
		}
	}()
	return g
}

// Take returns a slice of random bools of the given length.
func (g RandBoolGen) Take(count int) []bool {
	res := make([]bool, 0, count)
	for i := 0; i < count; i++ {
		res = append(res, <-g.ch)
	}
	return res
}

// Next returns a single random bool.
func (g RandBoolGen) Next() bool {
	return <-g.ch
}

// Stop closes all goroutines used during bool generation.
func (g RandBoolGen) Stop() {
	close(g.done)

	// The underlying generator might be stuck in the default select case trying to push a value into the channel,
	// so we need to make sure it is unstuck to be able to close the output channel
	<-g.ch
}
