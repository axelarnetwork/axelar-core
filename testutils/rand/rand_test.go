package rand

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestIntGen(t *testing.T) {
	t.Run("return given number of elements", intTake10)
	t.Run("limit range", between4And17Take43)
	t.Run("limit range and exclude number through where clause", between7And10Without8Take500)
	t.Run("multiple where clauses", between0And50GreaterThan25LesserThan30OrExactly45Take20)
}

func TestRandBoolGen(t *testing.T) {
	t.Run("return given number of elements", boolTake13)
	t.Run("correct ratio", ratioConvergesToOneEigth)
}

func intTake10(t *testing.T) {
	g := PInt64Gen()

	nums := g.Take(10)
	assert.Equal(t, 10, len(nums))
}

func between4And17Take43(t *testing.T) {
	g := I64GenBetween(4, 17)

	nums := g.Take(43)

	assert.Equal(t, 43, len(nums))
	for _, n := range nums {
		assert.True(t, n >= 4)
		assert.True(t, n < 17)
	}
}

func between7And10Without8Take500(t *testing.T) {
	g := I64GenBetween(7, 10).Where(func(n int64) bool { return n != 8 })

	nums := g.Take(500)

	assert.Equal(t, 500, len(nums))
	for _, n := range nums {
		assert.True(t, n >= 7)
		assert.True(t, n < 10)
		assert.NotEqual(t, 8, n)
	}
}

func between0And50GreaterThan25LesserThan30OrExactly45Take20(t *testing.T) {
	g := I64GenBetween(0, 50).
		Where(func(n int64) bool { return n > 25 }).
		Where(func(n int64) bool { return n < 30 || n == 45 })

	nums := g.Take(20)
	assert.Equal(t, 20, len(nums))
	for _, n := range nums {
		assert.True(t, n > 25)
		assert.True(t, n < 30 || n == 45)
	}
}

func boolTake13(t *testing.T) {
	g := Bools(0.5)

	assert.Equal(t, 13, len(g.Take(13)))
}

// this test is testing distribution sampling, so in very rare cases the test can fail (outlier sampling)
func ratioConvergesToOneEigth(t *testing.T) {
	expectedRatio := 1.0 / 8
	g := Bools(expectedRatio)

	actualRatio := calcRatio(g.Take(50000))
	assert.InEpsilon(t, expectedRatio, actualRatio, 0.05)
}

func calcRatio(values []bool) float64 {
	ratio := 0.0
	for _, val := range values {
		if val {
			ratio++
		}
	}
	ratio /= float64(len(values))
	return ratio
}

func TestRandDistinctStringGen_Take_DifferentLengths(t *testing.T) {
	g := Strings(1, 100).Distinct()

	previous := map[string]bool{}
	for _, s := range g.Take(1000) {
		previous[s] = true
	}
	assert.Len(t, previous, 1000)
}

func TestRandDistinctStringGen_Take_SameLength(t *testing.T) {
	g := Strings(10, 10).Distinct()

	previous := map[string]struct{}{}
	for _, s := range g.Take(1000) {
		assert.Len(t, s, 10)
		previous[s] = struct{}{}
	}
	assert.Len(t, previous, 1000)
}

func TestRandomValidator(t *testing.T) {
	address1 := ValAddr()
	address2, err := sdk.ValAddressFromBech32(address1.String())
	assert.NoError(t, err)
	assert.Equal(t, address1, address2)
}

func TestRandomAddress(t *testing.T) {
	address1 := AccAddr()
	address2, err := sdk.AccAddressFromBech32(address1.String())
	assert.NoError(t, err)
	assert.Equal(t, address1, address2)
}
