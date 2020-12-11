package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandIntGen(t *testing.T) {
	t.Run("return given number of elements", intTake10)
	t.Run("limit range", between4And17Take43)
	t.Run("limit range and exclude number through where clause", between7And10Without8Take500)
	t.Run("multiple where clauses", between0And50GreaterThan25LesserThan30OrExactly45Take20)
	t.Run("close all channels when done", intStop)
}

func TestRandBoolGen(t *testing.T) {
	t.Run("return given number of elements", boolTake13)
	t.Run("correct ratio", ratioConvergesToOneEigth)
	t.Run("close all channels when done", boolStop)
}

func intTake10(t *testing.T) {
	g := RandInts()
	defer g.Stop()

	nums := g.Take(10)
	assert.Equal(t, 10, len(nums))
}

func between4And17Take43(t *testing.T) {
	g := RandIntsBetween(4, 17)
	defer g.Stop()

	nums := g.Take(43)

	assert.Equal(t, 43, len(nums))
	for _, n := range nums {
		assert.True(t, n >= 4)
		assert.True(t, n < 17)
	}
}

func between7And10Without8Take500(t *testing.T) {
	g := RandIntsBetween(7, 10).Where(func(n int64) bool { return n != 8 })
	defer g.Stop()

	nums := g.Take(500)

	assert.Equal(t, 500, len(nums))
	for _, n := range nums {
		assert.True(t, n >= 7)
		assert.True(t, n < 10)
		assert.NotEqual(t, 8, n)
	}
}

func between0And50GreaterThan25LesserThan30OrExactly45Take20(t *testing.T) {
	g := RandIntsBetween(0, 50).
		Where(func(n int64) bool { return n > 25 }).
		Where(func(n int64) bool { return n < 30 || n == 45 })
	defer g.Stop()

	nums := g.Take(20)
	assert.Equal(t, 20, len(nums))
	for _, n := range nums {
		assert.True(t, n > 25)
		assert.True(t, n < 30 || n == 45)
	}
}

func intStop(t *testing.T) {
	g1 := RandInts()
	g2 := g1.Where(func(_ int64) bool { return true })
	g3 := g2.Where(func(_ int64) bool { return true })
	g4 := g3.Where(func(_ int64) bool { return true })
	g5 := g4.Where(func(_ int64) bool { return true })
	g6 := g5.Where(func(_ int64) bool { return true })
	g7 := g6.Where(func(_ int64) bool { return true })
	g8 := g7.Where(func(_ int64) bool { return true })
	_ = g8.Take(10)

	g8.Stop()

	_, ok := <-g1.ch
	assert.False(t, ok)
	_, ok = <-g2.ch
	assert.False(t, ok)
	_, ok = <-g3.ch
	assert.False(t, ok)
	_, ok = <-g4.ch
	assert.False(t, ok)
	_, ok = <-g5.ch
	assert.False(t, ok)
	_, ok = <-g6.ch
	assert.False(t, ok)
	_, ok = <-g7.ch
	assert.False(t, ok)
	_, ok = <-g8.ch
	assert.False(t, ok)
}

func boolTake13(t *testing.T) {
	g := RandBools(0.5)
	defer g.Stop()

	assert.Equal(t, 13, len(g.Take(13)))
}

// this test is testing distribution sampling, so in very rare cases the test can fail (outlier sampling)
func ratioConvergesToOneEigth(t *testing.T) {
	expectedRatio := 1.0 / 8
	g := RandBools(expectedRatio)
	defer g.Stop()

	actualRatio := calcRatio(g.Take(50000))
	assert.InEpsilon(t, expectedRatio, actualRatio, 0.05)
}

func calcRatio(values []bool) float64 {
	ratio := 0.0
	for _, val := range values {
		if val {
			ratio += 1
		}
	}
	ratio /= float64(len(values))
	return ratio
}

func boolStop(t *testing.T) {
	g := RandBools(0.3246)
	g.Stop()

	_, ok := <-g.ch
	assert.False(t, ok)
}

func TestRandDistinctStringGen_Take_DifferentLengths(t *testing.T) {
	g := RandStrings(1, 100).Distinct()
	defer g.Stop()

	previous := map[string]struct{}{}
	for _, s := range g.Take(1000) {
		previous[s] = struct{}{}
	}
	assert.Len(t, previous, 1000)
}

func TestRandDistinctStringGen_Take_SameLength(t *testing.T) {
	g := RandStrings(10, 10).Distinct()
	defer g.Stop()

	previous := map[string]struct{}{}
	for _, s := range g.Take(1000) {
		assert.Len(t, s, 10)
		previous[s] = struct{}{}
	}
	assert.Len(t, previous, 1000)
}
