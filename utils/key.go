package utils

import (
	"bytes"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Key represents a store key to interact with the NormalizedKVStore
type Key interface {
	// AsKey returns the byte representation of the key. If given, uses a delimiter string to separate prefixes
	AsKey(delimiter ...string) []byte
	Prepend(prefix Key) Key
	Append(key Key) Key
	Equals(key Key) bool
}

// NormalizedKVStore is a wrapper around the cosmos-sdk KVStore to provide more safety regarding key management and better ease-of-use
type NormalizedKVStore struct {
	sdk.KVStore
	cdc codec.BinaryMarshaler
}

// NewNormalizedStore returns a new NormalizedKVStore
func NewNormalizedStore(store sdk.KVStore, cdc codec.BinaryMarshaler) NormalizedKVStore {
	return NormalizedKVStore{
		KVStore: store,
		cdc:     cdc,
	}
}

// Set marshals the value and stores it under the given key
func (store NormalizedKVStore) Set(key Key, value codec.ProtoMarshaler) {
	store.KVStore.Set(key.AsKey(), store.cdc.MustMarshalBinaryLengthPrefixed(value))
}

// SetRaw stores the value under the given key
func (store NormalizedKVStore) SetRaw(key Key, value []byte) {
	store.KVStore.Set(key.AsKey(), value)
}

// Get unmarshals the raw bytes stored under the given key into the value object. Returns true if the key exists.
func (store NormalizedKVStore) Get(key Key, value codec.ProtoMarshaler) bool {
	bz := store.KVStore.Get(key.AsKey())
	if bz == nil {
		return false
	}
	store.cdc.MustUnmarshalBinaryLengthPrefixed(bz, value)
	return true
}

// GetRaw returns the raw bytes stored under the given key. Returns nil with key does not exist.
func (store NormalizedKVStore) GetRaw(key Key) []byte {
	return store.KVStore.Get(key.AsKey())
}

// Has returns true if the key exists.
func (store NormalizedKVStore) Has(key Key) bool {
	return store.KVStore.Has(key.AsKey())
}

// Delete deletes the value stored under the given key, if it exists
func (store NormalizedKVStore) Delete(key Key) {
	store.KVStore.Delete(key.AsKey())
}

type key struct {
	prefix Key
	key    []byte
}

// KeyFromStr applies the optional string transformations to the given key in sequence and returns a structured key
func KeyFromStr(k string, stringTransformations ...func(string) string) Key {
	for _, transform := range stringTransformations {
		k = transform(k)
	}
	return key{
		prefix: nil,
		key:    []byte(k),
	}
}

// LowerCaseKey returns a key with the input converted to lower case
func LowerCaseKey(k string) Key {
	return KeyFromStr(k, strings.ToLower)
}

// KeyFromBz returns a structured key
func KeyFromBz(k []byte) Key {
	return key{
		prefix: nil,
		key:    k,
	}
}

// AsKey returns the byte representation of the key. If given, uses a delimiter string to separate prefixes (default is "_")
func (k key) AsKey(delimiter ...string) []byte {
	if len(delimiter) == 0 {
		return k.asKey("_")
	}
	return k.asKey(delimiter[0])
}

func (k key) asKey(delimiter string) []byte {
	if k.prefix != nil {
		prefix := k.prefix.AsKey()
		delim := []byte(delimiter)
		compKey := make([]byte, 0, len(prefix)+len(delim)+len(k.key))
		return append(append(append(compKey, prefix...), delim...), k.key...)
	}
	return k.key
}

// Prepend prepends the given prefix to the key
func (k key) Prepend(prefix Key) Key {
	if k.prefix != nil {
		k.prefix = k.prefix.Prepend(prefix)
	} else {
		k.prefix = prefix
	}

	return k
}

// Append appends the given key to this key
func (k key) Append(key Key) Key {
	return key.Prepend(k)
}

// Equals compares two keys for equality
func (k key) Equals(other Key) bool {
	return bytes.Equal(k.AsKey(), other.AsKey())
}
