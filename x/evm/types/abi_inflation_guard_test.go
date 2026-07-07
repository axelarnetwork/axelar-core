package types

import (
	"encoding/binary"
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/utils/funcs"
)

// maxArgCost is the decoded-value budget the guard is tested against.
// Most messages are capped by 1MB
const maxArgCost = 1024 * 1024

// abi_inflation_guard.go ports several unexported helpers verbatim from
// go-ethereum's accounts/abi. These tests pin the upstream source of each ported
// helper, so an upstream change produces a readable diff telling us to re-check
// the port (and the surrounding walker) before bumping go-ethereum.
func TestGoEthereumABISourceUnchanged(t *testing.T) {
	for _, tc := range []struct {
		// anchor is an exported function defined in the same upstream file as the
		// ported helper; reflecting on it resolves that file's on-disk path
		// without depending on the go toolchain being available.
		anchor   interface{}
		funcName string
		want     string
	}{
		{abi.NewType, "isDynamicType", `func isDynamicType(t Type) bool {
	if t.T == TupleTy {
		for _, elem := range t.TupleElems {
			if isDynamicType(*elem) {
				return true
			}
		}
		return false
	}
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy || (t.T == ArrayTy && isDynamicType(*t.Elem))
}`},
		{abi.NewType, "requiresLengthPrefix", `func (t Type) requiresLengthPrefix() bool {
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy
}`},
		{abi.ReadInteger, "lengthPrefixPointsTo", `func lengthPrefixPointsTo(index int, output []byte) (start int, length int, err error) {
	bigOffsetEnd := new(big.Int).SetBytes(output[index : index+32])
	bigOffsetEnd.Add(bigOffsetEnd, common.Big32)
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
}`},
		{abi.ReadInteger, "tuplePointsTo", `func tuplePointsTo(index int, output []byte) (start int, err error) {
	offset := new(big.Int).SetBytes(output[index : index+32])
	outputLen := big.NewInt(int64(len(output)))

	if offset.Cmp(outputLen) > 0 {
		return 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)", offset, outputLen)
	}
	if offset.BitLen() > 63 {
		return 0, fmt.Errorf("abi offset larger than int64: %v", offset)
	}
	return int(offset.Uint64()), nil
}`},
		{abi.ReadInteger, "toGoType", `func toGoType(index int, t Type, output []byte) (interface{}, error) {
	if index+32 > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %d require %d", len(output), index+32)
	}

	var (
		returnOutput  []byte
		begin, length int
		err           error
	)

	// if we require a length prefix, find the beginning word and size returned.
	if t.requiresLengthPrefix() {
		begin, length, err = lengthPrefixPointsTo(index, output)
		if err != nil {
			return nil, err
		}
	} else {
		returnOutput = output[index : index+32]
	}

	switch t.T {
	case TupleTy:
		if isDynamicType(t) {
			begin, err := tuplePointsTo(index, output)
			if err != nil {
				return nil, err
			}
			return forTupleUnpack(t, output[begin:])
		}
		return forTupleUnpack(t, output[index:])
	case SliceTy:
		return forEachUnpack(t, output[begin:], 0, length)
	case ArrayTy:
		if isDynamicType(*t.Elem) {
			offset := binary.BigEndian.Uint64(returnOutput[len(returnOutput)-8:])
			if offset > uint64(len(output)) {
				return nil, fmt.Errorf("abi: toGoType offset greater than output length: offset: %d, len(output): %d", offset, len(output))
			}
			return forEachUnpack(t, output[offset:], 0, t.Size)
		}
		return forEachUnpack(t, output[index:], 0, t.Size)
	case StringTy: // variable arrays are written at the end of the return bytes
		return string(output[begin : begin+length]), nil
	case IntTy, UintTy:
		return ReadInteger(t, returnOutput)
	case BoolTy:
		return readBool(returnOutput)
	case AddressTy:
		return common.BytesToAddress(returnOutput), nil
	case HashTy:
		return common.BytesToHash(returnOutput), nil
	case BytesTy:
		return output[begin : begin+length], nil
	case FixedBytesTy:
		return ReadFixedBytes(t, returnOutput)
	case FunctionTy:
		return readFunctionType(t, returnOutput)
	default:
		return nil, fmt.Errorf("abi: unknown type %v", t.T)
	}
}`},
		{abi.ReadInteger, "forTupleUnpack", `func forTupleUnpack(t Type, output []byte) (interface{}, error) {
	retval := reflect.New(t.GetType()).Elem()
	virtualArgs := 0
	for index, elem := range t.TupleElems {
		marshalledValue, err := toGoType((index+virtualArgs)*32, *elem, output)
		if err != nil {
			return nil, err
		}
		if elem.T == ArrayTy && !isDynamicType(*elem) {
			// If we have a static array, like [3]uint256, these are coded as
			// just like uint256,uint256,uint256.
			// This means that we need to add two 'virtual' arguments when
			// we count the index from now on.
			//
			// Array values nested multiple levels deep are also encoded inline:
			// [2][3]uint256: uint256,uint256,uint256,uint256,uint256,uint256
			//
			// Calculate the full array size to get the correct offset for the next argument.
			// Decrement it by 1, as the normal index increment is still applied.
			virtualArgs += getTypeSize(*elem)/32 - 1
		} else if elem.T == TupleTy && !isDynamicType(*elem) {
			// If we have a static tuple, like (uint256, bool, uint256), these are
			// coded as just like uint256,bool,uint256
			virtualArgs += getTypeSize(*elem)/32 - 1
		}
		retval.Field(index).Set(reflect.ValueOf(marshalledValue))
	}
	return retval.Interface(), nil
}`},
		{abi.ReadInteger, "forEachUnpack", `func forEachUnpack(t Type, output []byte, start, size int) (interface{}, error) {
	if size < 0 {
		return nil, fmt.Errorf("cannot marshal input to array, size is negative (%d)", size)
	}
	if start+32*size > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal into go array: offset %d would go over slice boundary (len=%d)", len(output), start+32*size)
	}

	// this value will become our slice or our array, depending on the type
	var refSlice reflect.Value

	switch t.T {
	case SliceTy:
		// declare our slice
		refSlice = reflect.MakeSlice(t.GetType(), size, size)
	case ArrayTy:
		// declare our array
		refSlice = reflect.New(t.GetType()).Elem()
	default:
		return nil, errors.New("abi: invalid type in array/slice unpacking stage")
	}

	// Arrays have packed elements, resulting in longer unpack steps.
	// Slices have just 32 bytes per element (pointing to the contents).
	elemSize := getTypeSize(*t.Elem)

	for i, j := start, 0; j < size; i, j = i+elemSize, j+1 {
		inter, err := toGoType(i, *t.Elem, output)
		if err != nil {
			return nil, err
		}

		// append the item to our reflect slice
		refSlice.Index(j).Set(reflect.ValueOf(inter))
	}

	// return the interface
	return refSlice.Interface(), nil
}`},
		{abi.NewType, "getTypeSize", `func getTypeSize(t Type) int {
	if t.T == ArrayTy && !isDynamicType(*t.Elem) {
		// Recursively calculate type size if it is a nested array
		if t.Elem.T == ArrayTy || t.Elem.T == TupleTy {
			return t.Size * getTypeSize(*t.Elem)
		}
		return t.Size * 32
	} else if t.T == TupleTy && !isDynamicType(t) {
		total := 0
		for _, elem := range t.TupleElems {
			total += getTypeSize(*elem)
		}
		return total
	}
	return 32
}`},
	} {
		anchor := runtime.FuncForPC(reflect.ValueOf(tc.anchor).Pointer())
		path, _ := anchor.FileLine(anchor.Entry())
		contents := funcs.Must(os.ReadFile(path))

		got := extractFunctionSource(t, string(contents), tc.funcName)
		assert.Equalf(t, tc.want, got,
			"go-ethereum %s changed; re-diff the verbatim port in abi_inflation_guard.go and update the pinned source", tc.funcName)
	}
}

// extractFunctionSource returns the verbatim source of the named top-level
// function declaration (from the func keyword to its closing brace).
func extractFunctionSource(t *testing.T, content, funcName string) string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	require.NoError(t, err)

	var src string
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			src = content[fn.Pos()-1 : fn.End()-1]
			return false
		}
		return true
	})
	require.NotEmptyf(t, src, "function %q not found in upstream source", funcName)
	return src
}

func typeOf(t *testing.T, typeStr string) abi.Type {
	t.Helper()
	typ, err := abi.NewType(typeStr, typeStr, nil)
	require.NoError(t, err)
	return typ
}

func argsOf(t *testing.T, typeStr string) abi.Arguments {
	t.Helper()
	return abi.Arguments{{Type: typeOf(t, typeStr)}}
}

func tupleOf(t *testing.T, components ...abi.ArgumentMarshaling) abi.Arguments {
	t.Helper()
	typ, err := abi.NewType("tuple", "", components)
	require.NoError(t, err)
	return abi.Arguments{{Type: typ}}
}

// word returns the 32-byte big-endian encoding of v.
func word(v uint64) []byte {
	b := make([]byte, 32)
	binary.BigEndian.PutUint64(b[24:], v)
	return b
}

// hugeWord returns a 32-byte word whose value has more than 63 bits set.
func hugeWord() []byte {
	b := make([]byte, 32)
	b[0] = 0xFF
	return b
}

func assertFits(t *testing.T, args abi.Arguments, payload []byte, maxCost int) {
	t.Helper()
	err := ABIInflationGuard(args, payload, maxCost)
	assert.NoError(t, err)
}

func assertTooLarge(t *testing.T, args abi.Arguments, payload []byte, maxCost int) {
	t.Helper()
	err := ABIInflationGuard(args, payload, maxCost)
	assert.ErrorIs(t, err, ErrTooLarge)
}

func assertMalformed(t *testing.T, args abi.Arguments, payload []byte, maxCost int) {
	t.Helper()
	err := ABIInflationGuard(args, payload, maxCost)
	assert.ErrorIs(t, err, ErrMalformed)
}

func assertBoundary(t *testing.T, args abi.Arguments, payload []byte, words int) {
	t.Helper()
	size := words * 32
	assertFits(t, args, payload, maxArgCost)
	assertFits(t, args, payload, size)
	assertTooLarge(t, args, payload, size-1)
}

func TestABIInflationGuard(t *testing.T) {
	// --- valid payloads: accepted, with the exact boundary pinned ---

	t.Run("scalars", func(t *testing.T) {
		args := abi.Arguments{
			{Type: typeOf(t, "uint256")},
			{Type: typeOf(t, "bool")},
			{Type: typeOf(t, "address")},
			{Type: typeOf(t, "bytes4")},
		}
		packed := funcs.Must(args.Pack(big.NewInt(5), true, [20]byte{}, [4]byte{1, 2, 3, 4}))
		assertBoundary(t, args, packed, 4) // one word each
	})

	t.Run("string and bytes", func(t *testing.T) {
		args := abi.Arguments{{Type: typeOf(t, "string")}, {Type: typeOf(t, "bytes")}}
		packed := funcs.Must(args.Pack("hello world", []byte{1, 2, 3, 4, 5}))
		assertBoundary(t, args, packed, 2) // each content rounds up to one word
	})

	t.Run("fixed array", func(t *testing.T) {
		args := argsOf(t, "uint256[2]")
		packed := funcs.Must(args.Pack([2]*big.Int{big.NewInt(7), big.NewInt(9)}))
		assertBoundary(t, args, packed, 2)
		_, err := args.Unpack(packed) // the payload is genuinely decodable
		require.NoError(t, err)
	})

	t.Run("dynamic array", func(t *testing.T) {
		args := argsOf(t, "uint256[]")
		packed := funcs.Must(args.Pack([]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}))
		assertBoundary(t, args, packed, 3) // 3 static elements
	})

	t.Run("empty dynamic array never inflates", func(t *testing.T) {
		args := argsOf(t, "uint256[]")
		packed := funcs.Must(args.Pack([]*big.Int{}))
		assertFits(t, args, packed, maxArgCost)
		assertFits(t, args, packed, 0) // zero-cost, no reject threshold
	})

	t.Run("fixed array of dynamic elements", func(t *testing.T) {
		args := argsOf(t, "string[2]")
		packed := funcs.Must(args.Pack([2]string{"a", "b"}))
		assertBoundary(t, args, packed, 4) // 2 offset heads + 1 word per string
	})

	t.Run("static tuple", func(t *testing.T) {
		args := tupleOf(t,
			abi.ArgumentMarshaling{Name: "A", Type: "uint256"},
			abi.ArgumentMarshaling{Name: "B", Type: "bool"},
		)
		packed := funcs.Must(args.Pack(struct {
			A *big.Int
			B bool
		}{big.NewInt(1), true}))
		assertBoundary(t, args, packed, 2)
	})

	t.Run("dynamic tuple with a nested static array field", func(t *testing.T) {
		args := tupleOf(t,
			abi.ArgumentMarshaling{Name: "A", Type: "uint256[2]"},
			abi.ArgumentMarshaling{Name: "B", Type: "string"},
		)
		packed := funcs.Must(args.Pack(struct {
			A [2]*big.Int
			B string
		}{[2]*big.Int{big.NewInt(1), big.NewInt(2)}, "x"}))
		assertBoundary(t, args, packed, 3) // 2 (static array) + 1 (string)
	})

	t.Run("accumulates across arguments", func(t *testing.T) {
		args := abi.Arguments{{Type: typeOf(t, "uint256[]")}, {Type: typeOf(t, "uint256[]")}}
		packed := funcs.Must(args.Pack(
			[]*big.Int{big.NewInt(1), big.NewInt(2)},
			[]*big.Int{big.NewInt(3), big.NewInt(4), big.NewInt(5)},
		))
		assertBoundary(t, args, packed, 5) // 2 + 3 elements
	})

	// --- DoS vectors: rejected at any budget (the overflow-safety the guard exists for) ---

	// vector 1: 32*size overflows go-ethereum's int64 bounds check.
	t.Run("rejects fixed array whose size overflows the word count", func(t *testing.T) {
		args := argsOf(t, "uint256[288230376151711744]")
		assertTooLarge(t, args, make([]byte, 32), maxArgCost)
	})

	// vector 2: nested static array, 2^40 leaves from 40 brackets.
	t.Run("rejects deeply nested static array", func(t *testing.T) {
		typeStr := "uint256"
		for i := 0; i < 40; i++ {
			typeStr += "[2]"
		}
		assertTooLarge(t, argsOf(t, typeStr), make([]byte, 32), maxArgCost)
	})

	// vector 3: 50^6 leaves.
	t.Run("rejects wide nested static array", func(t *testing.T) {
		args := argsOf(t, "uint256[50][50][50][50][50][50]")
		assertTooLarge(t, args, make([]byte, 32), maxArgCost)
	})

	// vector 4: C3 — nested [0] product. The [0] leaf zeroes output size, but iteration
	// happens at the parent [400] levels, so the product still trips the budget.
	t.Run("rejects nested [0] product", func(t *testing.T) {
		args := argsOf(t, "uint256[0][400][400][400]")
		assertTooLarge(t, args, make([]byte, 32), maxArgCost)
	})

	// vector 5: deep [1] nest (2^depth via the depth-charge) then a wide [32768] outer; 100 brackets.
	t.Run("rejects deep [1] nest then wide outer", func(t *testing.T) {
		typeStr := "uint256"
		for i := 0; i < 99; i++ {
			typeStr += "[1]"
		}
		typeStr += "[32768]"
		assertTooLarge(t, argsOf(t, typeStr), make([]byte, 32), maxArgCost)
	})

	// vector 6: empty-tuple array. The empty tuple is 0 words, but the each==0->1 floor makes
	// each [400] dimension count, so the product trips the budget.
	t.Run("rejects empty-tuple array", func(t *testing.T) {
		args := argsOf(t, "tuple[400][400][400]")
		assertTooLarge(t, args, make([]byte, 32), maxArgCost)
	})

	// fixed array of dynamic elements with a declared size beyond the budget.
	t.Run("rejects huge fixed array of dynamic elements", func(t *testing.T) {
		args := argsOf(t, "string[40000]")
		assertTooLarge(t, args, make([]byte, 32), maxArgCost)
	})

	// an oversized field propagates errTooLarge out of the tuple walk.
	t.Run("rejects a tuple with an oversized field", func(t *testing.T) {
		args := tupleOf(t, abi.ArgumentMarshaling{Name: "A", Type: "uint256[288230376151711744]"})
		assertTooLarge(t, args, make([]byte, 32), maxArgCost)
	})

	// dynamic amplification: a small payload whose N outer slice entries all
	// alias the SAME inner uint256[] of length N materialises N*N elements.
	t.Run("rejects dynamic array amplification from a small payload", func(t *testing.T) {
		const n = 512 // n*n = 262144 words >> maxArgSize/32 budget

		var payload []byte
		payload = append(payload, word(32)...) // offset to outer array
		payload = append(payload, word(n)...)  // outer length
		innerOff := uint64(32 * n)             // every entry aliases the one inner array
		for i := 0; i < n; i++ {
			payload = append(payload, word(innerOff)...)
		}
		payload = append(payload, word(n)...) // inner length
		for i := 0; i < n; i++ {
			payload = append(payload, word(1)...)
		}

		require.Less(t, len(payload), maxArgCost)
		assertTooLarge(t, argsOf(t, "uint256[][]"), payload, maxArgCost)
	})

	// --- malformed payloads: reported as errMalformed (not inflation) so the
	//     caller rejects them and the decoder is never handed an unwalkable payload ---

	t.Run("too-short payload is malformed", func(t *testing.T) {
		assertMalformed(t, argsOf(t, "uint256"), make([]byte, 10), maxArgCost)
	})

	t.Run("string with a bad length prefix is malformed", func(t *testing.T) {
		assertMalformed(t, argsOf(t, "string"), word(9999), maxArgCost)
	})

	t.Run("slice with an out-of-bounds offset is malformed", func(t *testing.T) {
		assertMalformed(t, argsOf(t, "uint256[]"), word(9999), maxArgCost)
	})

	t.Run("slice with an oversized length prefix is malformed", func(t *testing.T) {
		// offset 32 is valid, but the length word there has more than 63 bits.
		payload := append(word(32), hugeWord()...)
		assertMalformed(t, argsOf(t, "uint256[]"), payload, maxArgCost)
	})

	t.Run("slice whose length overruns the payload is malformed", func(t *testing.T) {
		// offset 32 is valid and length 100 is small, but 64+100*32 exceeds the payload.
		payload := append(word(32), word(100)...)
		assertMalformed(t, argsOf(t, "uint256[]"), payload, maxArgCost)
	})

	t.Run("fixed-array-of-dynamic with a bad array offset is malformed", func(t *testing.T) {
		assertMalformed(t, argsOf(t, "string[2]"), word(9999), maxArgCost)
	})

	t.Run("malformed element inside a dynamic array is malformed", func(t *testing.T) {
		// string[2]: the array offset (32) is valid, but element 0's string offset
		// is out of bounds - the error must propagate out of the element loop.
		payload := append(append(word(32), word(9999)...), word(0)...)
		assertMalformed(t, argsOf(t, "string[2]"), payload, maxArgCost)
	})

	t.Run("dynamic tuple with a bad offset is malformed", func(t *testing.T) {
		args := tupleOf(t, abi.ArgumentMarshaling{Name: "A", Type: "string"})
		assertMalformed(t, args, word(9999), maxArgCost)
	})

	t.Run("unknown type yields an error", func(t *testing.T) {
		// not producible by abi.NewType, but the defensive default branch is real.
		args := abi.Arguments{{Type: abi.Type{T: 250}}}
		err := ABIInflationGuard(args, make([]byte, 32), maxArgCost)
		require.Error(t, err)
	})

	t.Run("skips indexed arguments", func(t *testing.T) {
		args := abi.Arguments{
			{Type: typeOf(t, "uint256[288230376151711744]"), Indexed: true}, // skipped
			{Type: typeOf(t, "uint256")},
		}
		packed := funcs.Must(abi.Arguments{{Type: typeOf(t, "uint256")}}.Pack(big.NewInt(1)))
		assertFits(t, args, packed, maxArgCost)
	})
}
