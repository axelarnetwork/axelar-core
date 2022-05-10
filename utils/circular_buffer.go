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
	if m.isFull() && m.isLTMaxSize() {
		m.increaseBufferSize()
	}

	prevValue := m.CumulativeValue[m.Index]
	m.Index = m.addToIndex(1)
	m.CumulativeValue[m.Index] = prevValue + uint64(value)
}

// Count returns the cumulative value for the most recent given window
func (m CircularBuffer) Count(windowRange int) uint64 {
	if windowRange >= int(m.MaxSize) {
		panic(fmt.Errorf("window range to large"))
	}

	if windowRange >= len(m.CumulativeValue) {
		return m.CumulativeValue[m.Index]
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
	return int(m.Index)+1 == len(m.CumulativeValue)
}

func (m CircularBuffer) isLTMaxSize() bool {
	return len(m.CumulativeValue) < int(m.MaxSize)
}

// double buffer size until it reaches max size. If max size is not a power of 2 limit the last increase to max size
func (m *CircularBuffer) increaseBufferSize() {
	newBuffer := make([]uint64, math.Min(len(m.CumulativeValue)<<1, int(m.MaxSize)))
	copy(newBuffer, m.CumulativeValue)

	m.CumulativeValue = newBuffer
}
