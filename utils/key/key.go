package key

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/constraints"

	"github.com/axelarnetwork/utils/convert"
)

var usedPrefixes = map[prefixID]struct{}{}

type prefixID struct {
	namespace string
	prefix    uint
}

// RegisterStaticKey registers a static key particle for the data store to ensure uniqueness per namespace.
// Panics if two prefixes with the same value for the same namespace are registered.
func RegisterStaticKey(namespace string, prefix uint) Key {
	id := prefixID{
		namespace: namespace,
		prefix:    prefix,
	}
	if _, ok := usedPrefixes[id]; ok {
		panic(fmt.Sprintf("prefix key %d for namespace %s already registered", prefix, namespace))
	}

	usedPrefixes[id] = struct{}{}

	return FromUInt(prefix)
}

// DefaultDelimiter represents the default delimiter used for the KV store keys when concatenating them together
const DefaultDelimiter = "_"

// Key provides a type safe way to interact with the store
type Key interface {
	Append(Key) Key
	Bytes() []byte
	String() string
}

type basicKey struct {
	particles [][]byte
}

func (k basicKey) Append(suffix Key) Key {
	var bz [][]byte
	switch suffix := suffix.(type) {
	case basicKey:
		bz = suffix.particles
	default:
		bz = [][]byte{suffix.Bytes()}
	}
	return basicKey{
		particles: append(k.particles, bz...),
	}
}

func (k basicKey) Bytes() []byte {
	return bytes.Join(k.particles, []byte(DefaultDelimiter))
}

func (k basicKey) String() string {
	return string(k.Bytes())
}

// FromBz creates a new Key from bytes
// Deprecated: use FromBzHashed instead
func FromBz(key []byte) Key {
	return &basicKey{particles: [][]byte{key}}
}

// FromStr creates a new Key from a string
// Deprecated: use FromStrHashed instead
func FromStr(key string) Key {
	return FromBz([]byte(strings.ToLower(key)))
}

// From creates a new Key from a fmt.Stringer interface
// Deprecated: use FromHashed instead
func From(key fmt.Stringer) Key {
	return FromStr(key.String())
}

// FromUInt creates a new Key from any unsigned integer type
func FromUInt[T constraints.Unsigned](key T) Key {
	return &basicKey{particles: [][]byte{convert.IntToBytes(key)}}
}

// FromBzHashed creates a new Key from bytes
func FromBzHashed(key []byte) Key {
	hash := sha3.Sum256(key)
	return basicKey{particles: [][]byte{hash[:]}}
}

// FromStrHashed creates a new Key from a string
func FromStrHashed(key string) Key {
	return FromBzHashed([]byte(strings.ToLower(key)))
}

// FromHashed creates a new Key from a fmt.Stringer interface
func FromHashed(key fmt.Stringer) Key {
	return FromStrHashed(key.String())
}
