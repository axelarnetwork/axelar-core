package utils

import (
	"math/big"
)

// NewBitmap is the constructor for Bitmap
func NewBitmap() Bitmap {
	return Bitmap{
		Bits: []byte{},
	}
}

func (m Bitmap) getBits() *big.Int {
	return new(big.Int).SetBytes(m.Bits)
}

// Add adds a new bit to the Bitmap
func (m *Bitmap) Add(bit bool) *Bitmap {
	bits := m.getBits()
	bits = bits.Lsh(bits, 1)

	if bit {
		bits = bits.Add(bits, big.NewInt(1))
	}

	m.Bits = bits.Bytes()
	m.TrueCountCache = nil

	return m
}

// CountTrue returns the number of 1's in the given range
func (m *Bitmap) CountTrue(bitCount uint64) uint64 {
	if m.TrueCountCache != nil && m.TrueCountCache.BitCount == bitCount {
		return m.TrueCountCache.Result
	}

	m.TrueCountCache = nil

	result := uint64(0)
	bits := m.getBits()
	bitLen := bits.BitLen()

	for i := 0; uint64(i) < bitCount && i < bitLen; i++ {
		result += uint64(bits.Bit(i))
	}

	return result
}

// CountFalse returns the number of 0's in the given range
func (m *Bitmap) CountFalse(bitCount uint64) uint64 {
	return bitCount - m.CountTrue(bitCount)
}
