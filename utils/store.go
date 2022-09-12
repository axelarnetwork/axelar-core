package utils

import (
	"bytes"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"golang.org/x/exp/constraints"

	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/utils/convert"
)

//go:generate moq -pkg mock -out ./mock/store.go . ValidatedProtoMarshaler

// DefaultDelimiter represents the default delimiter used for the KV store keys
const DefaultDelimiter = "_"

// Key represents a store key to interact with the KVStore
// Deprecated: use key.Key instead
type Key interface {
	// AsKey returns the byte representation of the key. If given, uses a delimiter string to separate prefixes
	AsKey(delimiter ...string) []byte
	Prepend(prefix Key) Key
	Append(key Key) Key
}

// StringKey extends the Key interface for simplified appending and prepending
// Deprecated: use key.Key instead
type StringKey interface {
	Key
	AppendStr(key string, stringTransformations ...func(string) string) StringKey
	PrependStr(key string, stringTransformations ...func(string) string) StringKey
}

// KVStore is a wrapper around the cosmos-sdk KVStore to provide more safety regarding key management and better ease-of-use
type KVStore struct {
	sdk.KVStore
	cdc codec.BinaryCodec
}

// NewNormalizedStore returns a new KVStore
func NewNormalizedStore(store sdk.KVStore, cdc codec.BinaryCodec) KVStore {
	return KVStore{
		KVStore: store,
		cdc:     cdc,
	}
}

// Set marshals the value and stores it under the given key
// Deprecated: use SetNewValidated instead
func (store KVStore) Set(key Key, value codec.ProtoMarshaler) {
	store.KVStore.Set(key.AsKey(), store.cdc.MustMarshalLengthPrefixed(value))
}

// SetRawNew stores the value under the given key
func (store KVStore) SetRawNew(k key.Key, value []byte) {
	store.KVStore.Set(k.Bytes(), value)
}

// SetNewValidated marshals the value and stores it under the given key if it is valid
func (store KVStore) SetNewValidated(k key.Key, value ValidatedProtoMarshaler) error {
	if err := value.ValidateBasic(); err != nil {
		return err
	}

	store.KVStore.Set(k.Bytes(), store.cdc.MustMarshalLengthPrefixed(value))
	return nil
}

// SetRaw stores the value under the given key
// Deprecated: use SetRawNew instead
func (store KVStore) SetRaw(key Key, value []byte) {
	store.KVStore.Set(key.AsKey(), value)
}

// Get unmarshals the raw bytes stored under the given key into the value object. Returns true if the key exists.
// Deprecated: use GetNew instead
func (store KVStore) Get(key Key, value codec.ProtoMarshaler) bool {
	value.Reset()

	bz := store.KVStore.Get(key.AsKey())
	if bz == nil {
		return false
	}
	store.cdc.MustUnmarshalLengthPrefixed(bz, value)
	return true
}

// GetNew unmarshals the raw bytes stored under the given key into the value object. Returns true if the key exists.
func (store KVStore) GetNew(key key.Key, value codec.ProtoMarshaler) bool {
	value.Reset()

	bz := store.KVStore.Get(key.Bytes())
	if bz == nil {
		return false
	}
	store.cdc.MustUnmarshalLengthPrefixed(bz, value)
	return true
}

// GetRawNew returns the raw bytes stored under the given key. Returns nil with key does not exist.
func (store KVStore) GetRawNew(key key.Key) []byte {
	return store.KVStore.Get(key.Bytes())
}

// GetRaw returns the raw bytes stored under the given key. Returns nil with key does not exist.
// Deprecated: use GetRawNew instead
func (store KVStore) GetRaw(key Key) []byte {
	return store.KVStore.Get(key.AsKey())
}

// Has returns true if the key exists
// Deprecated: use HasNew instead
func (store KVStore) Has(key Key) bool {
	return store.KVStore.Has(key.AsKey())
}

// HasNew deletes the value stored under the given key, if it exists
func (store KVStore) HasNew(k key.Key) bool {
	return store.KVStore.Has(k.Bytes())
}

// Delete deletes the value stored under the given key, if it exists
// Deprecated: use DeleteNew instead
func (store KVStore) Delete(key Key) {
	store.KVStore.Delete(key.AsKey())
}

// DeleteNew deletes the value stored under the given key, if it exists
func (store KVStore) DeleteNew(k key.Key) {
	store.KVStore.Delete(k.Bytes())
}

// DeleteRaw deletes the value stored under the given raw key, if it exists
func (store KVStore) DeleteRaw(key []byte) {
	store.KVStore.Delete(key)
}

// Iterator returns an Iterator that can handle a structured Key
func (store KVStore) Iterator(prefix Key) Iterator {
	iter := sdk.KVStorePrefixIterator(store.KVStore, append(prefix.AsKey(), []byte(DefaultDelimiter)...))
	return iterator{Iterator: iter, cdc: store.cdc}
}

// IteratorNew returns an Iterator that can handle a structured Key
func (store KVStore) IteratorNew(prefix key.Key) Iterator {
	iter := sdk.KVStorePrefixIterator(store.KVStore, append(prefix.Bytes(), []byte(DefaultDelimiter)...))
	return iterator{Iterator: iter, cdc: store.cdc}
}

// ReverseIterator returns an Iterator that can handle a structured Key and
// interate reversely
func (store KVStore) ReverseIterator(prefix Key) Iterator {
	iter := sdk.KVStoreReversePrefixIterator(store.KVStore, append(prefix.AsKey(), []byte(DefaultDelimiter)...))
	return iterator{Iterator: iter, cdc: store.cdc}
}

// Iterator is an easier and safer to use sdk.Iterator extension
type Iterator interface {
	sdk.Iterator
	UnmarshalValue(marshaler codec.ProtoMarshaler)
	GetKey() Key
}

type iterator struct {
	sdk.Iterator
	cdc codec.BinaryCodec
}

// UnmarshalValue returns the value marshalled into the given type
func (i iterator) UnmarshalValue(value codec.ProtoMarshaler) {
	value.Reset()
	i.cdc.MustUnmarshalLengthPrefixed(i.Value(), value)
}

// GetKey returns the key of the current iterator value
func (i iterator) GetKey() Key {
	return KeyFromBz(i.Key())
}

type basicKey struct {
	prefix Key
	key    []byte
}

// KeyFromStr applies the optional string transformations to the given key in sequence and returns a structured key
func KeyFromStr(k string, stringTransformations ...func(string) string) StringKey {
	for _, transform := range stringTransformations {
		k = transform(k)
	}
	return basicKey{
		prefix: nil,
		key:    []byte(k),
	}
}

// LowerCaseKey returns a key with the input converted to lower case
func LowerCaseKey(k string) StringKey {
	return KeyFromStr(k, strings.ToLower)
}

// KeyFromBz returns a structured key
func KeyFromBz(k []byte) StringKey {
	return basicKey{
		prefix: nil,
		key:    k,
	}
}

// KeyFromInt returns a structured key
func KeyFromInt[T constraints.Integer](k T) StringKey {
	return KeyFromBz(convert.IntToBytes(k))
}

// AsKey returns the byte representation of the key. If given, uses a delimiter string to separate prefixes (default is "_")
func (k basicKey) AsKey(delimiter ...string) []byte {
	if len(delimiter) == 0 {
		return k.asKey(DefaultDelimiter)
	}
	return k.asKey(delimiter[0])
}

func (k basicKey) asKey(delimiter string) []byte {
	if k.prefix != nil {
		prefix := k.prefix.AsKey(delimiter)
		delim := []byte(delimiter)
		compKey := make([]byte, 0, len(prefix)+len(delim)+len(k.key))
		return append(append(append(compKey, prefix...), delim...), k.key...)
	}
	return k.key
}

// Prepend prepends the given prefix to the key
func (k basicKey) Prepend(prefix Key) Key {
	if k.prefix != nil {
		k.prefix = k.prefix.Prepend(prefix)
	} else {
		k.prefix = prefix
	}

	return k
}

// PrependStr prepends the given string to this key
func (k basicKey) PrependStr(prefix string, stringTransformations ...func(string) string) StringKey {
	return k.Prepend(KeyFromStr(prefix, stringTransformations...)).(StringKey)
}

// Append appends the given key to this key
func (k basicKey) Append(key Key) Key {
	return key.Prepend(k)
}

// AppendStr appends the given string to this key
func (k basicKey) AppendStr(key string, stringTransformations ...func(string) string) StringKey {
	return KeyFromStr(key, stringTransformations...).Prepend(k).(StringKey)
}

// Equals compares two keys for equality
func (k basicKey) Equals(other Key) bool {
	return bytes.Equal(k.AsKey(), other.AsKey())
}

// CloseLogError closes the given iterator and logs if an error is returned
func CloseLogError(iter sdk.Iterator, logger log.Logger) {
	err := iter.Close()
	if err != nil {
		logger.Error(sdkerrors.Wrap(err, "failed to close kv store iterator").Error())
	}
}
