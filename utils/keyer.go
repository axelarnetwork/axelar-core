package utils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keyer represents a store key to interact with the NormalizedKVStore
type Keyer interface {
	AsKey() []byte
	WithPrefix(prefix string) Keyer
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
func (store NormalizedKVStore) Set(key Keyer, value codec.ProtoMarshaler) {
	store.KVStore.Set(key.AsKey(), store.cdc.MustMarshalBinaryLengthPrefixed(value))
}

// SetRaw stores the value under the given key
func (store NormalizedKVStore) SetRaw(key Keyer, value []byte) {
	store.KVStore.Set(key.AsKey(), value)
}

// Get unmarshals the raw bytes stored under the given key into the value object. Returns true if the key exists.
func (store NormalizedKVStore) Get(key Keyer, value codec.ProtoMarshaler) bool {
	bz := store.KVStore.Get(key.AsKey())
	if bz == nil {
		return false
	}
	store.cdc.MustUnmarshalBinaryLengthPrefixed(bz, value)
	return true
}

// GetRaw returns the raw bytes stored under the given key. Returns nil with key does not exist.
func (store NormalizedKVStore) GetRaw(key Keyer) []byte {
	return store.KVStore.Get(key.AsKey())
}

// Delete deletes the value stored under the given key, if it exists
func (store NormalizedKVStore) Delete(key Keyer) {
	store.KVStore.Delete(key.AsKey())
}

var _ Keyer = LowerCaseKey("")

// LowerCaseKey wraps around a key string to enable case insensitive comparisons
type LowerCaseKey string

// ToLowerCaseKey converts the string representation of the given object into a lower case key
func ToLowerCaseKey(key fmt.Stringer) LowerCaseKey {
	return LowerCaseKey(key.String())
}

// AsKey returns the byte representation of the key
func (k LowerCaseKey) AsKey() []byte {
	return []byte(strings.ToLower(string(k)))
}

// WithPrefix prepends the given prefix to the key
func (k LowerCaseKey) WithPrefix(prefix string) Keyer {
	return LowerCaseKey(prefix + "_" + string(k))
}

// Equals compares two keys for equality
func (k LowerCaseKey) Equals(other Keyer) {
	bytes.Equal(k.AsKey(), other.AsKey())
}
