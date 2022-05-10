package utils_test

import (
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
	}).Repeat(1))
}
