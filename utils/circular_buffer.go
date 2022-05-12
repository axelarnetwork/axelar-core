package utils

import (
	"fmt"

	"github.com/axelarnetwork/utils/math"
)

// NewCircularBuffer is the constructor of CircularBuffer
func NewCircularBuffer(maxSize int) *CircularBuffer {
	return &CircularBuffer{
		CumulativeValue: make([]uint64, 32),
		Index:           0,
		MaxSize:         int32(maxSize),
	}
}

// Add appends a new value into the CircularBuffer
func (m *CircularBuffer) Add(value uint32) {
	if m.isGTMaxSize() {
		m.shrink()
	}
	if m.isFull() && m.isLTMaxSize() {
		m.grow()
	}

	prevValue := m.CumulativeValue[m.Index]
	m.Index = m.addToIndex(1)
	m.CumulativeValue[m.Index] = prevValue + uint64(value)
}

func (m CircularBuffer) isGTMaxSize() bool {
	return len(m.CumulativeValue) > int(m.MaxSize)
}

func (m *CircularBuffer) shrink() {
	newBuffer := make([]uint64, int(m.MaxSize))

	for i := 0; i < len(newBuffer); i++ {
		newBuffer[len(newBuffer)-1-i] = m.CumulativeValue[m.addToIndex(int32(-i))]
	}

	m.Index = int32(len(newBuffer) - 1)
	m.CumulativeValue = newBuffer
}

// Count returns the cumulative value for the most recent given window
func (m CircularBuffer) Count(windowRange int) uint64 {
	if windowRange >= int(m.MaxSize) {
		panic(fmt.Errorf("window range to large"))
	}

	if windowRange >= len(m.CumulativeValue) {
		return m.CumulativeValue[m.Index] - m.CumulativeValue[m.addToIndex(1)]
	}

	return m.CumulativeValue[m.Index] - m.CumulativeValue[m.addToIndex(int32(-windowRange))]
}

func (m CircularBuffer) addToIndex(i int32) int32 {
	index := m.Index + i
	length := int32(len(m.CumulativeValue))
	index = (index + length) % length

	return index
}

func (m CircularBuffer) isFull() bool {
	return int(m.Index)+1 == len(m.CumulativeValue) || m.CumulativeValue[m.addToIndex(1)] != 0
}

func (m CircularBuffer) isLTMaxSize() bool {
	return len(m.CumulativeValue) < int(m.MaxSize)
}

// double buffer size until it reaches max size. If max size is not a power of 2 limit the last increase to max size
func (m *CircularBuffer) grow() {
	newBuffer := make([]uint64, math.Min(len(m.CumulativeValue)<<1, int(m.MaxSize)))

	// there is no information about the count outside the buffer range, so when the new buffer gets padded with zeroes
	// the oldest value also needs to be reset to zero,
	// otherwise windows larger than the old buffer size would produce a wrong count
	zeroValue := m.CumulativeValue[m.addToIndex(1)]
	for i := 0; i < len(m.CumulativeValue); i++ {
		newBuffer[i] = m.CumulativeValue[m.addToIndex(1+int32(i))] - zeroValue
	}

	m.Index = int32(len(m.CumulativeValue) - 1)
	m.CumulativeValue = newBuffer
}

// SetMaxSize sets the max size of the buffer to the given value.
// The buffer size gets updated accordingly the next time a value is added.
func (m *CircularBuffer) SetMaxSize(size int) {
	m.MaxSize = int32(size)
}
