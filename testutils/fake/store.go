package fake

import (
	"bytes"
	"io"
	"sort"
	"sync"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake/interfaces"
	"github.com/axelarnetwork/axelar-core/testutils/fake/interfaces/mock"
)

// MultiStore is a simple multistore used for testing
type MultiStore struct {
	kvstore map[string]interfaces.KVStore
	*mock.MultiStoreMock
}

// NewMultiStore returns a new Multistore instance used for testing
func NewMultiStore() sdk.MultiStore {
	ms := MultiStore{
		kvstore:        map[string]interfaces.KVStore{},
		MultiStoreMock: &mock.MultiStoreMock{},
	}
	ms.GetKVStoreFunc = func(storeKey types.StoreKey) types.KVStore {
		if store, ok := ms.kvstore[storeKey.String()]; ok {
			return store
		}
		store := NewTestKVStore()
		ms.kvstore[storeKey.String()] = store
		return store

	}
	return ms
}

// TestKVStore is a kv store for testing
type TestKVStore struct {
	mutex *sync.RWMutex
	store map[string][]byte
}

// NewTestKVStore returns a new kv store instance for testing
func NewTestKVStore() interfaces.KVStore {
	return TestKVStore{
		mutex: &sync.RWMutex{},
		store: map[string][]byte{},
	}
}

// GetStoreType is not implemented
func (t TestKVStore) GetStoreType() sdk.StoreType {
	panic("implement me")
}

// CacheWrap is not implemented
func (t TestKVStore) CacheWrap() sdk.CacheWrap {
	panic("implement me")
}

// CacheWrapWithTrace is not implemented
func (t TestKVStore) CacheWrapWithTrace(_ io.Writer, _ sdk.TraceContext) sdk.CacheWrap {
	panic("implement me")
}

// CacheWrapWithListeners is not implemented
func (t TestKVStore) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	panic("implement me")
}

// Get returns the value of the given key, nil if it does not exist
func (t TestKVStore) Get(key []byte) []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	val, ok := t.store[string(key)]

	if !ok {
		return nil
	}
	return val
}

// Has checks if an entry for the given key exists
func (t TestKVStore) Has(key []byte) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	_, ok := t.store[string(key)]
	return ok
}

// Set stores the given key value pair
func (t TestKVStore) Set(key, value []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.store[string(key)] = value
}

// Delete deletes a key if it exists
func (t TestKVStore) Delete(key []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	delete(t.store, string(key))
}

// Iterator returns an interator over the given key domain
func (t TestKVStore) Iterator(start, end []byte) sdk.Iterator {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	return newMockIterator(start, end, t.store)
}

// ReverseIterator returns an iterator that iterates over all keys in the given domain in reverse order
func (t TestKVStore) ReverseIterator(start, end []byte) sdk.Iterator {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	iter := newMockIterator(start, end, t.store)

	// reverse the order of the iterator, which is returned already
	// sorted in ascending order
	for i, j := 0, len(iter.keys)-1; i < j; i, j = i+1, j-1 {
		iter.keys[i], iter.keys[j] = iter.keys[j], iter.keys[i]
		iter.values[i], iter.values[j] = iter.values[j], iter.values[i]

	}

	iter.start = end
	iter.end = start

	return iter
}

// fake iterator
type mockIterator struct {
	keys       [][]byte
	values     [][]byte
	index      int
	start, end []byte
}

func newMockIterator(start, end []byte, content map[string][]byte) *mockIterator {

	keys := make([][]byte, 0)

	// select the keys according to the specified domain
	for k := range content {

		b := []byte(k)

		if (start == nil && end == nil) || (bytes.Compare(b, start) >= 0 && bytes.Compare(b, end) < 0) {

			// make sure data is a copy so that there is no concurrent writing
			temp := make([]byte, len(k))
			copy(temp, k)
			keys = append(keys, temp)
		}
	}

	// Sort the keys in ascending order
	sort.Slice(keys, func(i, j int) bool {

		return bytes.Compare(keys[i], keys[j]) < 0

	})

	// With the keys chosen and sorted, we can now populate the slice of values
	values := make([][]byte, len(keys))

	for i := 0; i < len(keys); i++ {

		// make sure data is a copy so that there is no concurrent writing
		value := content[string(keys[i])]
		temp := make([]byte, len(value))
		copy(temp, value)

		values[i] = temp
	}

	return &mockIterator{
		keys:   keys,
		values: values,
		index:  0,
		start:  start,
		end:    end,
	}
}

// Domain returns the key domain of the iterator.
// The start & end (exclusive) limits to iterate over.
// If end < start, then the Iterator goes in reverse order.
//
// A domain of ([]byte{12, 13}, []byte{12, 14}) will iterate
// over anything with the prefix []byte{12, 13}.
//
// The smallest key is the empty byte array []byte{} - see BeginningKey().
// The largest key is the nil byte array []byte(nil) - see EndingKey().
// CONTRACT: start, end readonly []byte
func (mi mockIterator) Domain() (start []byte, end []byte) {
	return mi.start, mi.end
}

// Valid returns whether the current position is valid.
// Once invalid, an Iterator is forever invalid.
func (mi mockIterator) Valid() bool {
	return mi.index < len(mi.keys)
}

// Next moves the iterator to the next sequential key in the database, as
// defined by order of iteration.
// If Valid returns false, this method will panic.
func (mi *mockIterator) Next() {
	mi.index++
}

// Key returns the key of the cursor.
// If Valid returns false, this method will panic.
// CONTRACT: key readonly []byte
func (mi mockIterator) Key() (key []byte) {
	if !mi.Valid() {
		panic("Iterator position out of bounds")
	}

	return mi.keys[mi.index]
}

// Value returns the value of the cursor.
// If Valid returns false, this method will panic.
// CONTRACT: value readonly []byte
func (mi mockIterator) Value() (value []byte) {
	if !mi.Valid() {
		panic("Iterator position out of bounds")
	}

	return mi.values[mi.index]
}

func (mi mockIterator) Error() error {
	return nil
}

// Close releases the Iterator.
func (mi mockIterator) Close() error {
	return nil
}
