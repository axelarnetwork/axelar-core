package rand

import (
	"math/rand"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"golang.org/x/text/unicode/norm"

	rand2 "github.com/axelarnetwork/utils/test/rand"
)

const (
	defaultDelimiter = "_"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// PosI64 returns a positive pseudo-random integer
func PosI64() int64 {
	x := rand.Int63()
	for x == 0 {
		x = rand.Int63()
	}
	return x
}

// UintBetween returns a random integer between lower (inclusive) and upper (exclusive).
// It panics if upper <= lower.
func UintBetween(lower sdk.Uint, upper sdk.Uint) sdk.Uint {
	return sdk.NewUint(uint64(rand.Int63n(upper.Sub(lower).BigInt().Int64()))).Add(lower)
}

// I64Between returns a random integer between lower (inclusive) and upper (exclusive).
// It panics if  upper <= lower.
func I64Between(lower int64, upper int64) int64 {
	return rand.Int63n(upper-lower) + lower
}

// NormalizeString normalizes a string as NFKC
func NormalizeString(str string) string {
	return norm.NFKC.String(str)
}

// NormalizedStr creates a random normalized string of the provided length
func NormalizedStr(len int) string {
	return NormalizedStrBetween(len, len+1)
}

// NormalizedStrBetween creates a random normalized string in the provided range (exclusive uper limit)
func NormalizedStrBetween(min, shorterThan int) string {
	return strings.ReplaceAll(NormalizeString(StrBetween(min, shorterThan)), defaultDelimiter, "-")
}

// I64Gen represents an random integer generator to generate a sequence of integers with the same properties.
// Call Stop when done so dangling goroutines can be cleaned up.
type I64Gen struct {
	gen func() int64
}

// PInt64Gen returns a random integer generator for positive integers.
func PInt64Gen() I64Gen {
	return I64Gen{gen: PosI64}
}

// I64GenBetween returns a random integer generator for numbers between lower (inclusive) and upper (exclusive).
// It panics if  upper <= lower.
func I64GenBetween(lower int64, upper int64) I64Gen {
	return I64Gen{gen: func() int64 { return I64Between(lower, upper) }}
}

// Where restricts the output of the underlying generator to adhere to the predicate.
// If the predicate is not satisfiable the Take function will deadlock.
func (g I64Gen) Where(predicate func(i int64) bool) I64Gen {
	newGen := func() int64 {
		n := g.Next()
		for !predicate(n) {
			n = g.Next()
		}
		return n
	}
	return I64Gen{gen: newGen}
}

// Take returns a slice of random integers of the given length.
func (g I64Gen) Take(count int) []int64 {
	nums := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		nums = append(nums, g.Next())
	}
	return nums
}

// Next returns a single random integer.
func (g I64Gen) Next() int64 {
	return g.gen()
}

// Bytes returns a random slice of bytes of the specified length
func Bytes(len int) []byte {
	bz := make([]byte, len)
	for i, b := range I64GenBetween(0, 256).Take(len) {
		bz[i] = byte(b)
	}
	return bz
}

// BytesBetween returns a random byte slice of random length in the given limits (upper exclusive)
func BytesBetween(lower int, upper int) []byte {
	len := int(I64Between(int64(lower), int64(upper)))
	bz := make([]byte, len)
	for i, b := range I64GenBetween(0, 256).Take(len) {
		bz[i] = byte(b)
	}
	return bz
}

// BoolGen represents an random bool generator.
// Call Stop when done so dangling goroutines can be cleaned up.
type BoolGen struct {
	ch    chan bool
	ratio float64
}

// Bools returns a random bool generator that adheres to the given ratio of true to false values.
func Bools(ratio float64) BoolGen {
	return BoolGen{ch: make(chan bool), ratio: ratio}
}

// Take returns a slice of random bools of the given length.
func (g BoolGen) Take(count int) []bool {
	res := make([]bool, 0, count)
	for i := 0; i < count; i++ {
		res = append(res, g.Next())
	}
	return res
}

// Next returns a single random bool.
func (g BoolGen) Next() bool {
	return rand.Float64() < g.ratio
}

// DistrGen represents a probability distribution that can be sampled
type DistrGen struct {
	total  int64
	states []int64
}

// Distr generates a new probability distribution with n states of random probability
func Distr(n int) DistrGen {
	if n < 1 {
		panic("at least one state necessary")
	}
	gen := &DistrGen{}
	// the larger the resolution the higher the potential deviation of probabilities between states
	var resolution int64 = 10

	// Ensure the total stays in int64
	if resolution*int64(n) > math.MaxInt32 {
		panic("decrease either number of states or resolution")
	}
	for _, n := range I64GenBetween(1, resolution*int64(n)).Take(n) {
		gen.total += n
		gen.states = append(gen.states, gen.total)
	}
	return *gen
}

// Samples returns n samples drawn from the given distribution
func (g DistrGen) Samples(n int) []int {
	var samples []int
	for i := 0; i < n; i++ {
		samples = append(samples, g.Draw())
	}
	return samples
}

// Draw returns a single sample drawn from the given distribution
func (g DistrGen) Draw() int {
	return binSearch(g.states, I64Between(0, g.total))
}

func binSearch(a []int64, search int64) int {
	mid := len(a) / 2
	switch {
	case len(a) == 0:
		return -1 // not found
	case a[mid] > search:
		return binSearch(a[:mid], search)
	case a[mid] < search:
		return binSearch(a[mid+1:], search)
	default:
		return mid
	}
}

// StringGen represents an random string generator.
// Call Stop when done so dangling goroutines can be cleaned up.
type StringGen struct {
	lengthGen  I64Gen
	alphabet   []rune
	charPicker I64Gen
}

// Strings returns a random string generator that produces strings from the default alphabet of random length in the given limits (upper limit exclusive)
func Strings(minLength int, shorterThan int) StringGen {
	alphabet := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-.:")
	return StringGen{
		lengthGen:  I64GenBetween(int64(minLength), int64(shorterThan)),
		alphabet:   alphabet,
		charPicker: I64GenBetween(0, int64(len(alphabet))),
	}
}

// Denom returns a random denom string (max exclusive)
func Denom(min, max int) string {
	// first letter must be an ascii alphabet
	return Strings(1, 2).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")).Next() + Strings(min, max).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-/")).Next()
}

// HexStrings returns a random hex string generator that produces hex strings with given length
func HexStrings(length int) StringGen {
	alphabet := []rune("0123456789abcdef")

	return StringGen{
		lengthGen:  I64GenBetween(int64(length), int64(length+1)),
		alphabet:   alphabet,
		charPicker: I64GenBetween(0, int64(len(alphabet))),
	}
}

// StrBetween returns a random string of random length in the given limits (upper exclusive)
func StrBetween(minLength int, shorterThan int) string {
	g := Strings(minLength, shorterThan)
	return g.Next()
}

// Str returns a random string of given length
func Str(len int) string {
	return StrBetween(len, len+1)
}

// HexStr returns a random hex string of given length
func HexStr(len int) string {
	return HexStrings(len).Next()
}

// WithAlphabet returns a random string generator that produces strings from the given alphabet
func (g StringGen) WithAlphabet(alphabet []rune) StringGen {
	return StringGen{
		lengthGen:  g.lengthGen,
		alphabet:   alphabet,
		charPicker: I64GenBetween(0, int64(len(alphabet))),
	}
}

// Take returns a slice of random strings of the given length.
func (g StringGen) Take(count int) []string {
	res := make([]string, 0, count)
	for i := 0; i < count; i++ {
		res = append(res, g.Next())
	}
	return res
}

// Next returns a single random string.
func (g StringGen) Next() string {
	s := make([]rune, g.lengthGen.Next())
	for i := range s {
		s[i] = g.alphabet[g.charPicker.Next()]
	}
	return string(s)
}

// Distinct returns a new unique string
func (g StringGen) Distinct() DistinctStrGen {
	return DistinctStrGen{StringGen: g, previous: make(map[string]bool)}
}

// DistinctStrGen represents an random string generator which returns distinct strings.
// Call Stop when done so dangling goroutines can be cleaned up.
type DistinctStrGen struct {
	StringGen
	previous map[string]bool
}

// Take returns a slice of distinct random strings of the given length.
func (g DistinctStrGen) Take(count int) []string {
	res := make([]string, 0, count)
	for i := 0; i < count; i++ {
		res = append(res, g.Next())
	}
	return res
}

// Next returns a single random string that is distinct from all previously generated strings.
func (g DistinctStrGen) Next() string {
	for {
		s := g.StringGen.Next()
		if ok := g.previous[s]; ok {
			continue
		}
		g.previous[s] = true
		return s
	}
}

// ValAddr generates a random validator address
func ValAddr() sdk.ValAddress {
	return Bytes(address.Len)
}

// AccAddr generates a random cosmos address
func AccAddr() sdk.AccAddress {
	return Bytes(address.Len)
}

// Context generates a random Context data structure
func Context(store types.MultiStore) sdk.Context {
	ctx := sdk.NewContext(store, tmproto.Header{Height: PosI64(), Time: rand2.Time()}, false, log.TestingLogger()).
		WithHeaderHash(BytesBetween(1024, 101240)).
		WithBlockGasMeter(sdk.NewGasMeter(1000000))
	ctx.GasMeter().ConsumeGas(uint64(I64Between(1000, 1000000)), "test")
	return ctx
}

// Of returns a random item from the given slice
func Of[T any](items ...T) T {
	return items[I64Between(0, int64(len(items)))]
}
