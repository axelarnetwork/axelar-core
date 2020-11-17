package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandInts(t *testing.T) {
	t.Run("return given number of elements", take10)
	t.Run("limit range", between4And17Take43)
	t.Run("limit range and exclude number through where clause", between7And10Without8Take500)
	t.Run("multiple where clauses", between0And50GreaterThan25LesserThan30OrExactly45Take20)
	t.Run("close all channels when done", closeAllChannels)
}

func take10(t *testing.T) {
	nums := RandInts().Take(10)
	assert.Equal(t, 10, len(nums))
}

func between4And17Take43(t *testing.T) {
	nums := RandIntsBetween(4, 17).Take(43)

	assert.Equal(t, 43, len(nums))
	for _, n := range nums {
		assert.True(t, n >= 4)
		assert.True(t, n < 17)
	}
}

func between7And10Without8Take500(t *testing.T) {
	nums := RandIntsBetween(7, 10).Where(func(n int) bool { return n != 8 }).Take(500)

	assert.Equal(t, 500, len(nums))
	for _, n := range nums {
		assert.True(t, n >= 7)
		assert.True(t, n < 10)
		assert.NotEqual(t, 8, n)
	}
}

func between0And50GreaterThan25LesserThan30OrExactly45Take20(t *testing.T) {
	nums := RandIntsBetween(0, 50).
		Where(func(n int) bool { return n > 25 }).
		Where(func(n int) bool { return n < 30 || n == 45 }).
		Take(20)

	assert.Equal(t, 20, len(nums))
	for _, n := range nums {
		assert.True(t, n > 25)
		assert.True(t, n < 30 || n == 45)
	}
}

func closeAllChannels(t *testing.T) {
	g1 := RandInts()
	g2 := g1.Where(func(_ int) bool { return true })
	g3 := g2.Where(func(_ int) bool { return true })
	g4 := g3.Where(func(_ int) bool { return true })
	g5 := g4.Where(func(_ int) bool { return true })
	g6 := g5.Where(func(_ int) bool { return true })
	g7 := g6.Where(func(_ int) bool { return true })
	g8 := g7.Where(func(_ int) bool { return true })
	_ = g8.Take(10)

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
