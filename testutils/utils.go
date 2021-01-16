// Package testutils provides general purpose utility functions for unit/integration testing.
package testutils

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	ethereum "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/types"
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
	codec.RegisterCrypto(cdc)

	// Add new modules here so tests have access to marshalling the registered ethereum
	vote.RegisterCodec(cdc)
	bitcoin.RegisterCodec(cdc)
	tss.RegisterCodec(cdc)
	broadcast.RegisterCodec(cdc)
	snapshot.RegisterCodec(cdc)
	ethereum.RegisterCodec(cdc)

	return cdc
}

// RandIntBetween returns a random integer between lower (inclusive) and upper (exclusive).
// It panics if  upper <= lower.
func RandIntBetween(lower int64, upper int64) int64 {
	return rand.Int63n(upper-lower) + lower
}

// RandBytes resturns a random slice of bytes of the specified length
func RandBytes(len int) []byte {
	bz := make([]byte, len)
	gen := RandIntsBetween(0, 256)
	defer gen.Stop()
	for i, b := range gen.Take(len) {
		bz[i] = byte(b)
	}
	return bz
}

// RandIntGen represents an random integer generator to generate a sequence of integers with the same properties.
// Call Stop when done so dangling goroutines can be cleaned up.
type RandIntGen struct {
	ch      chan int64
	done    chan struct{}
	wrapped *RandIntGen
}

// RandInts returns a random integer generator for positive integers.
func RandInts() RandIntGen {
	return generateInt64(rand.Int63)
}

// RandIntsBetween returns a random integer generator for numbers between lower (inclusive) and upper (exclusive).
// It panics if  upper <= lower.
func RandIntsBetween(lower int64, upper int64) RandIntGen {
	return generateInt64(func() int64 { return rand.Int63n(upper-lower) + lower })
}

// Where restricts the output of the underlying generator to adhere to the predicate.
// If the predicate is not satisfiable the Take function will deadlock.
func (g RandIntGen) Where(predicate func(i int64) bool) RandIntGen {
	newGen := RandIntGen{ch: make(chan int64), wrapped: &g}
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
func (g RandIntGen) Take(count int) []int64 {
	nums := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		nums = append(nums, <-g.ch)
	}
	return nums
}

// Next returns a single random integer.
func (g RandIntGen) Next() int64 {
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

func generateInt64(generator func() int64) RandIntGen {
	g := RandIntGen{ch: make(chan int64), done: make(chan struct{}), wrapped: nil}
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

// RandStringGen represents an random string generator.
// Call Stop when done so dangling goroutines can be cleaned up.
type RandStringGen struct {
	ch         chan string
	done       chan struct{}
	lengthGen  RandIntGen
	alphabet   []rune
	charPicker RandIntGen
}

// RandStrings returns a random string generator that produces strings of random length in the given limits (inclusive)
func RandStrings(minLength int, maxLength int) RandStringGen {
	alphabet := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.")
	g := RandStringGen{
		ch:         make(chan string),
		done:       make(chan struct{}),
		lengthGen:  RandIntsBetween(int64(minLength), int64(maxLength+1)),
		alphabet:   alphabet,
		charPicker: RandIntsBetween(0, int64(len(alphabet))),
	}
	go func() {
		for {
			select {
			case <-g.done:
				close(g.ch)
				return
			default:
				s := make([]rune, g.lengthGen.Next())
				for i := range s {
					s[i] = g.alphabet[g.charPicker.Next()]
				}
				g.ch <- string(s)
			}
		}
	}()
	return g
}

// RandString returns a random string of random length in the given limits (inclusive)
func RandString(len int) string {
	g := RandStrings(len, len)
	defer g.Stop()
	return g.Next()
}

// Take returns a slice of random strings of the given length.
func (g RandStringGen) Take(count int) []string {
	res := make([]string, 0, count)
	for i := 0; i < count; i++ {
		res = append(res, <-g.ch)
	}
	return res
}

// Next returns a single random string.
func (g RandStringGen) Next() string {
	return <-g.ch
}

// Stop closes all goroutines used during string generation.
func (g RandStringGen) Stop() {
	close(g.done)

	// The underlying generator might be stuck in the default select case trying to push a value into the channel,
	// so we need to make sure it is unstuck to be able to close the output channel
	<-g.ch
	g.charPicker.Stop()
	g.lengthGen.Stop()
}

// Distinct returns a new unique string
func (g RandStringGen) Distinct() RandDistinctStringGen {
	return RandDistinctStringGen{RandStringGen: g, previous: make(map[string]struct{})}
}

// RandDistinctStringGen represents an random string generator which returns distinct strings.
// Call Stop when done so dangling goroutines can be cleaned up.
type RandDistinctStringGen struct {
	RandStringGen
	previous map[string]struct{}
}

// Take returns a slice of distinct random strings of the given length.
func (g RandDistinctStringGen) Take(count int) []string {
	res := make([]string, 0, count)
	for i := 0; i < count; i++ {
		for {
			s := <-g.ch
			if _, ok := g.previous[s]; !ok {
				res = append(res, s)
				g.previous[s] = struct{}{}
				break
			}
		}
	}
	return res
}

// Next returns a single random string that is distinct from all previously generated strings.
func (g RandDistinctStringGen) Next() string {
	for {
		s := <-g.ch
		if _, ok := g.previous[s]; !ok {
			g.previous[s] = struct{}{}
			return s
		}
	}
}
