package utils

func NewCircularBuffer(maxSize int) *CircularBuffer {
	return &CircularBuffer{
		CumulativeValue: make([]uint64, 32),
		Index:           0,
		MaxSize:         int32(maxSize),
	}
}

func (m *CircularBuffer) Add(value uint32) {
	if m.bufferIsFull() && m.bufferLTMaxSize() {
		m.doubleBufferSize()
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

func (m *CircularBuffer) doubleBufferSize() {
	newBuffer := make([]uint64, len(m.CumulativeValue)<<1)
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
