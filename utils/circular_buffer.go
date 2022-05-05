package utils

import "github.com/axelarnetwork/utils/math"

func NewCircularBuffer(maxSize int) *CircularBuffer {
	return &CircularBuffer{
		CumulativeValue: make([]uint64, 32),
		Index:           0,
		MaxSize:         int32(maxSize),
	}
}

func (m *CircularBuffer) Add(value uint32) {
	if m.bufferIsFull() && m.bufferLTMaxSize() {
		m.increaseBufferSize()
	}

	prevValue := m.CumulativeValue[m.Index]
	m.Index = m.addToIndex(1)
	m.CumulativeValue[m.Index] = prevValue + uint64(value)
}

func (m CircularBuffer) addToIndex(i int32) int32 {
	index := m.Index + i
	length := int32(len(m.CumulativeValue))
	index = (index + length) % length
	return index
}

func (m CircularBuffer) bufferIsFull() bool {
	return int(m.Index)+1 == len(m.CumulativeValue)
}

func (m CircularBuffer) bufferLTMaxSize() bool {
	return (len(m.CumulativeValue) << 1) <= int(m.MaxSize)
}

// double buffer size until it reaches max size. If max size is not a power of 2 limit the last increase to max size
func (m *CircularBuffer) increaseBufferSize() {
	newBuffer := make([]uint64, math.Min(len(m.CumulativeValue)<<1, int(m.MaxSize)))
	for i := 0; i < len(m.CumulativeValue); i++ {
		newBuffer[i] = m.CumulativeValue[i]
	}
	m.CumulativeValue = newBuffer
}

func (m CircularBuffer) Count(windowRange int) uint64 {
	if windowRange >= int(m.MaxSize) {
		panic("window range to large")
	}

	if windowRange >= len(m.CumulativeValue) {
		return m.CumulativeValue[m.Index]
	}

	return m.CumulativeValue[m.Index] - m.CumulativeValue[m.addToIndex(int32(-windowRange))]
}
