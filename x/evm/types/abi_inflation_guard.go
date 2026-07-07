package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	ErrTooLarge  = errors.New("abi: decoded value exceeds maximum allowed cost")
	ErrMalformed = errors.New("abi: malformed payload")
)

// ABIInflationGuard reports whether ABI-decoding the payload against the given
// argument types would exceed an inflation cost of maxCost. The cost captures
// not just the materialised memory but also the recursion/processing work the
// decoder would do (nesting depth and per-element overhead inflate it). This
// function walks the same length/offset prefixes that the abi.Arguments.Unpack
// decoder would follow but only sums each value's cost without allocating.
func ABIInflationGuard(arguments abi.Arguments, payload []byte, maxCost int) error {
	// the limit, and all cost counts, are measured in 32-byte words.
	limit := maxCost / 32

	total, index, virtualArgs := 0, 0, 0
	for _, arg := range arguments {
		if arg.Indexed {
			continue
		}

		words, err := walk((index+virtualArgs)*32, arg.Type, payload, limit-total)
		if err != nil {
			return err
		}

		total += words
		if total > limit {
			return ErrTooLarge
		}

		// static arrays/tuples are encoded inline, so the head index advances by
		// their full encoded word count.
		if (arg.Type.T == abi.ArrayTy || arg.Type.T == abi.TupleTy) && !isDynamicType(arg.Type) {
			virtualArgs += getTypeSize(arg.Type)/32 - 1
		}
		index++
	}

	return nil
}

// walk mirrors go-ethereum's toGoType (github.com/ethereum/go-ethereum/blob/master/accounts/abi/unpack.go#L224)
func walk(index int, t abi.Type, output []byte, limit int) (int, error) {
	if index+32 > len(output) {
		// Changed: generic malformed error.
		return 0, ErrMalformed
	}

	var (
		returnOutput  []byte
		begin, length int
		err           error
	)

	if requiresLengthPrefix(t) {
		begin, length, err = lengthPrefixPointsTo(index, output)
		if err != nil {
			// Changed: generic malformed error.
			return 0, ErrMalformed
		}
	} else {
		returnOutput = output[index : index+32]
	}

	switch t.T {
	case abi.TupleTy:
		if isDynamicType(t) {
			begin, err := tuplePointsTo(index, output)
			if err != nil {
				// Changed: generic malformed error.
				return 0, ErrMalformed
			}
			return walkTuple(t, output[begin:], limit)
		}
		return walkTuple(t, output[index:], limit)

	case abi.SliceTy:
		return walkElements(t, output[begin:], 0, length, limit)

	case abi.ArrayTy:
		if isDynamicType(*t.Elem) {
			offset := binary.BigEndian.Uint64(returnOutput[len(returnOutput)-8:])
			if offset > uint64(len(output)) {
				return 0, ErrMalformed
			}
			return walkElements(t, output[offset:], 0, t.Size, limit)
		}
		return walkElements(t, output[index:], 0, t.Size, limit)

	// Changed: calculate cost instead of returning data.
	case abi.StringTy, abi.BytesTy:
		return (length + 31) / 32, nil

	// Changed: calculate cost instead of returning data.
	case abi.IntTy, abi.UintTy, abi.BoolTy, abi.AddressTy, abi.HashTy, abi.FixedBytesTy, abi.FunctionTy:
		return 1, nil

	default:
		return 0, fmt.Errorf("abi: unknown type %d", t.T)
	}
}

// walkTuple mirrors go-ethereum's forTupleUnpack (github.com/ethereum/go-ethereum/blob/master/accounts/abi/unpack.go#L192)
func walkTuple(t abi.Type, output []byte, limit int) (int, error) {
	total, virtualArgs := 0, 0
	for index, elem := range t.TupleElems {
		// Changed: get cost instead of marshalledValue back.
		words, err := walk((index+virtualArgs)*32, *elem, output, limit-total)
		if err != nil {
			return 0, err
		}
		// Changed: calculate cost and error if too large.
		total += words
		if total > limit {
			return 0, ErrTooLarge
		}
		if elem.T == abi.ArrayTy && !isDynamicType(*elem) {
			virtualArgs += getTypeSize(*elem)/32 - 1
		} else if elem.T == abi.TupleTy && !isDynamicType(*elem) {
			virtualArgs += getTypeSize(*elem)/32 - 1
		}
	}
	return total, nil
}

// walkElements mirrors go-ethereum's forEachUnpack (github.com/ethereum/go-ethereum/blob/master/accounts/abi/unpack.go#L152)
func walkElements(t abi.Type, output []byte, start, size, limit int) (int, error) {
	if size < 0 {
		// Changed: generic malformed error.
		return 0, ErrMalformed
	}
	// Removed: a check that can overflow, but not needed as we do real counting.
	// Added: shortcut for when item size is zero.
	if t.T != abi.ArrayTy && size == 0 {
		return 0, nil
	}

	elem := *t.Elem

	// Added: shortcut that doesn't loop large fixed arrays.
	// Walk one type and multiply by size. Also overflow aware!
	if !isDynamicType(elem) {
		// Added: Each array dimension should inflate cost because depth adds process cost.
		if t.T == abi.ArrayTy && size < 2 {
			size = 2
		}
		each, err := walk(start, elem, output, limit)
		if err != nil {
			return 0, err
		}
		// Added: Each subarray inflates cost because depth adds process cost.
		if each == 0 {
			each = 1
		}
		if size > limit/each {
			return 0, ErrTooLarge
		}
		return size * each, nil
	}

	// Removed: the creation of refSlice.
	// Added: Check for size.
	total := size
	if total > limit {
		return 0, ErrTooLarge
	}

	// Removed: elemSize calculation. The dynamic type's head is always 32.
	// As the static type is done above, we can calculate here with 32.
	for i, j := start, 0; j < size; i, j = i+32, j+1 {
		words, err := walk(i, elem, output, limit-total)
		if err != nil {
			return 0, err
		}
		// Changed: do total checking instead of output generation.
		total += words
		if total > limit {
			return 0, ErrTooLarge
		}
	}
	// Changed: get cost back.
	return total, nil
}

// Exact copy of go-ethereum getTypeSize (github.com/ethereum/go-ethereum/blob/master/accounts/abi/type.go#L384)
func getTypeSize(t abi.Type) int {
	if t.T == abi.ArrayTy && !isDynamicType(*t.Elem) {
		// Recursively calculate type size if it is a nested array
		if t.Elem.T == abi.ArrayTy || t.Elem.T == abi.TupleTy {
			return t.Size * getTypeSize(*t.Elem)
		}
		return t.Size * 32
	} else if t.T == abi.TupleTy && !isDynamicType(t) {
		total := 0
		for _, elem := range t.TupleElems {
			total += getTypeSize(*elem)
		}
		return total
	}
	return 32
}

// Exact copy of go-ethereum requiresLengthPrefix (github.com/ethereum/go-ethereum/blob/master/accounts/abi/type.go#L353)
func requiresLengthPrefix(t abi.Type) bool {
	return t.T == abi.StringTy || t.T == abi.BytesTy || t.T == abi.SliceTy
}

// Exact copy of go-ethereum isDynamicType (github.com/ethereum/go-ethereum/blob/master/accounts/abi/type.go#L364)
func isDynamicType(t abi.Type) bool {
	if t.T == abi.TupleTy {
		for _, elem := range t.TupleElems {
			if isDynamicType(*elem) {
				return true
			}
		}
		return false
	}
	return t.T == abi.StringTy || t.T == abi.BytesTy || t.T == abi.SliceTy || (t.T == abi.ArrayTy && isDynamicType(*t.Elem))
}

// Exact copy of go-ethereum lengthPrefixPointsTo (github.com/ethereum/go-ethereum/blob/master/accounts/abi/unpack.go#L288)
func lengthPrefixPointsTo(index int, output []byte) (start int, length int, err error) {
	bigOffsetEnd := new(big.Int).SetBytes(output[index : index+32])
	bigOffsetEnd.Add(bigOffsetEnd, big.NewInt(32))
	outputLength := big.NewInt(int64(len(output)))

	if bigOffsetEnd.Cmp(outputLength) > 0 {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)", bigOffsetEnd, outputLength)
	}

	if bigOffsetEnd.BitLen() > 63 {
		return 0, 0, fmt.Errorf("abi offset larger than int64: %v", bigOffsetEnd)
	}

	offsetEnd := int(bigOffsetEnd.Uint64())
	lengthBig := new(big.Int).SetBytes(output[offsetEnd-32 : offsetEnd])

	totalSize := new(big.Int).Add(bigOffsetEnd, lengthBig)
	if totalSize.BitLen() > 63 {
		return 0, 0, fmt.Errorf("abi: length larger than int64: %v", totalSize)
	}

	if totalSize.Cmp(outputLength) > 0 {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %v require %v", outputLength, totalSize)
	}
	start = int(bigOffsetEnd.Uint64())
	length = int(lengthBig.Uint64())
	return
}

// Exact copy of go-ethereum tuplePointsTo (github.com/ethereum/go-ethereum/blob/master/accounts/abi/unpack.go#L318)
func tuplePointsTo(index int, output []byte) (start int, err error) {
	offset := new(big.Int).SetBytes(output[index : index+32])
	outputLen := big.NewInt(int64(len(output)))

	if offset.Cmp(outputLen) > 0 {
		return 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)", offset, outputLen)
	}
	if offset.BitLen() > 63 {
		return 0, fmt.Errorf("abi offset larger than int64: %v", offset)
	}
	return int(offset.Uint64()), nil
}
