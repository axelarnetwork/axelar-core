package utils

import (
	"math/big"
)

// var mask = sdk.NewInt(2).ToDec().Power(MaxBitmapLen - 1).RoundInt().BigInt()

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

	return m
}

// CountTrue returns the number of 1's in the given range
func (m Bitmap) CountTrue(bitCount uint) uint {
	result := uint(0)
	bits := m.getBits()
	bitLen := bits.BitLen()

	for i := 0; uint(i) < bitCount && i < bitLen; i++ {
		result += bits.Bit(i)
	}

	return result
}

// CountFalse returns the number of 0's in the given range
func (m Bitmap) CountFalse(bitCount uint) uint {
	return bitCount - m.CountTrue(bitCount)
}
