package utils_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	. "github.com/axelarnetwork/utils/test"
)

func TestCircularBuffer(t *testing.T) {
	t.Run("Add and Count", testutils.Func(func(t *testing.T) {
		var circularBuffer *utils.CircularBuffer
		var maxSize int
		var windowRange int

		givenNewCircularBitmap := Given("a new bitmap", func() {
			maxSize = int(rand.I64Between(10, 10000))
			circularBuffer = utils.NewCircularBuffer(maxSize)
		})

		givenNewCircularBitmap.
			When("window range >= max size", func() {
				windowRange = maxSize + int(rand.I64Between(0, 10))
			}).
			Then("should panic", func(t *testing.T) {
				assert.PanicsWithError(t, "window range to large", func() {
					circularBuffer.Count(windowRange)
				})
			}).
			Run(t)

		n := uint32(rand.I64Between(1, 100))
		total := uint32(0)
		givenNewCircularBitmap.
			When("added with 1...n", func() {
				for i := uint32(1); i <= n; i++ {
					circularBuffer.Add(i)
					total += i
				}
			}).
			Then("should count correctly", func(t *testing.T) {
				for windowRange := 0; windowRange < maxSize; windowRange++ {
					if windowRange <= int(n) {
						expected := uint32(0)
						for i := n; i > n-uint32(windowRange); i-- {
							expected += i
						}

						assert.EqualValues(t, expected, circularBuffer.Count(windowRange))
					} else {
						assert.EqualValues(t, total, circularBuffer.Count(windowRange))
					}
				}
			}).
			Run(t)

		n = uint32(rand.I64Between(1, 10))
		givenNewCircularBitmap.
			When("added with many n's", func() {
				for i := 0; i < maxSize*10; i++ {
					circularBuffer.Add(n)
				}
			}).
			Then("should count correctly", func(t *testing.T) {
				for windowRange := 0; windowRange < maxSize; windowRange++ {
					assert.EqualValues(t, int(n)*windowRange, circularBuffer.Count(windowRange))
				}
			}).
			Run(t)

		n = uint32(rand.I64Between(1, 10))
		bufferIsShrunkAfterBeingFilled := When("the buffer is full", func() {
			for i := 0; i < 2*maxSize; i++ {
				circularBuffer.Add(n)
			}
		}).
			When("max size is decreased", func() {
				circularBuffer.SetMaxSize(2)
			})

		givenNewCircularBitmap.
			When2(bufferIsShrunkAfterBeingFilled).
			Then("return correct count", func(t *testing.T) {
				assert.EqualValues(t, n, circularBuffer.Count(1))
			}).Run(t)

		givenNewCircularBitmap.
			When2(bufferIsShrunkAfterBeingFilled).
			When("adding another entry", func() {
				circularBuffer.Add(2 * n)
			}).
			Then("return correct count", func(t *testing.T) {
				assert.EqualValues(t, 2*n, circularBuffer.Count(1))
			}).Run(t)

		bufferIsFull := When("it is completely full", func() {
			circularBuffer.CumulativeValue[circularBuffer.Index] = math.MaxUint64 - uint64(circularBuffer.MaxSize)
			// increase buffer size to max size
			for i := int32(0); i < circularBuffer.MaxSize; i++ {
				circularBuffer.Add(1)
			}
		})
		givenNewCircularBitmap.
			When2(bufferIsFull).
			Then("do not overflow", func(t *testing.T) {
				circularBuffer.Add(1)
				assert.EqualValues(t, 5, circularBuffer.Count(5))
			}).Run(t)

		givenNewCircularBitmap.
			When2(bufferIsFull).
			When("buffer gets shrunk", func() {
				circularBuffer.SetMaxSize(10)
				circularBuffer.Add(1)
			}).
			When("buffer gets enlarged again", func() {
				circularBuffer.SetMaxSize(100)
				circularBuffer.Add(1)
			}).
			Then("return correct count", func(t *testing.T) {
				assert.EqualValues(t, 10, circularBuffer.Count(30))
			}).Run(t)

	}).Repeat(1))

}
