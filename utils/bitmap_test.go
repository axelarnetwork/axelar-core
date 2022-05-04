package utils_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	. "github.com/axelarnetwork/utils/test"
)

func TestBitmap(t *testing.T) {
	t.Run("Add, CountTrue and CountFalse", testutils.Func(func(t *testing.T) {
		var bitmap utils.Bitmap
		var trueCount, falseCount uint

		givenNewBitmap := Given("a new bitmap", func(_ *testing.T) { bitmap = utils.NewBitmap() })
		whenPushedWith := func(value bool, count uint) ThenStatement {
			return givenNewBitmap.When(fmt.Sprintf("pushed with %d %t's", count, value), func(_ *testing.T) {
				for i := uint(0); i < count; i++ {
					bitmap.Add(value)
				}
			})
		}

		trueCount = uint(rand.I64Between(1, 10000))
		whenPushedWith(true, trueCount).
			Then("should have correct number of true's", func(t *testing.T) {
				assert.Equal(t, trueCount-1, bitmap.CountTrue(trueCount-1))
				assert.Equal(t, trueCount, bitmap.CountTrue(trueCount))
				assert.Equal(t, trueCount, bitmap.CountTrue(trueCount+1))
			}).
			Run(t)

		falseCount = uint(rand.I64Between(1, 10000))
		whenPushedWith(false, falseCount).
			Then("should have correct number of false's", func(t *testing.T) {
				assert.Equal(t, falseCount-1, bitmap.CountFalse(falseCount-1))
				assert.Equal(t, falseCount, bitmap.CountFalse(falseCount))
				assert.Equal(t, falseCount+1, bitmap.CountFalse(falseCount+1))
			}).
			Run(t)

		total := uint(rand.I64Between(1, 10000))
		trueCount = uint(rand.I64Between(1, int64(total)))
		falseCount = total - trueCount
		givenNewBitmap.
			When(fmt.Sprintf("pushed with pushed with %d true's and %d false's", trueCount, falseCount), func(_ *testing.T) {
				tCount := trueCount
				fCount := falseCount

				for i := uint(0); i < total; i++ {
					var value bool

					switch {
					case tCount > 0 && fCount > 0:
						value = rand.Bools(0.5).Next()
					case tCount > 0:
						value = true
					case fCount > 0:
						value = false
					default:
						t.Fatalf("impossible")
					}

					if value {
						tCount--
					} else {
						fCount--
					}

					bitmap.Add(value)
				}
			}).
			Then("should have correct number of true's and false's", func(t *testing.T) {
				assert.Equal(t, trueCount, bitmap.CountTrue(total))
				assert.Equal(t, falseCount, bitmap.CountFalse(total))
			}).
			Run(t)
	}).Repeat(20))
}
