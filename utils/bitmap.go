package utils

// NewBitmap is the constructor for Bitmap
func NewBitmap(maxSize int) Bitmap {
	return Bitmap{
		TrueCountCache: NewCircularBuffer(maxSize),
	}
}

// Add adds a new bit to the Bitmap
func (m *Bitmap) Add(bit bool) *Bitmap {
	if bit {
		m.TrueCountCache.Add(1)
	} else {
		m.TrueCountCache.Add(0)
	}
	return m
}

// CountTrue returns the number of 1's in the given range
func (m Bitmap) CountTrue(bitCount int) uint64 {
	return m.TrueCountCache.Count(bitCount)
}

// CountFalse returns the number of 0's in the given range
func (m Bitmap) CountFalse(bitCount int) uint64 {
	return uint64(bitCount) - m.CountTrue(bitCount)
}
