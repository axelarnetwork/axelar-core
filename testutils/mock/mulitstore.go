package mock

import (
	"bytes"
	"io"
	"sort"
	"sync"

	sdkTypes "github.com/cosmos/cosmos-sdk/types"
)

func NewMultiStore() sdkTypes.MultiStore {
	return TestMultiStore{kvstore: make(map[string]sdkTypes.KVStore)}
}

type TestMultiStore struct {
	kvstore map[string]sdkTypes.KVStore
}

func (t TestMultiStore) GetStoreType() sdkTypes.StoreType {
	panic("implement me")
}

func (t TestMultiStore) CacheWrap() sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestMultiStore) CacheWrapWithTrace(_ io.Writer, _ sdkTypes.TraceContext) sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestMultiStore) CacheMultiStore() sdkTypes.CacheMultiStore {
	panic("implement me")
}

func (t TestMultiStore) CacheMultiStoreWithVersion(_ int64) (sdkTypes.CacheMultiStore, error) {
	panic("implement me")
}

func (t TestMultiStore) GetStore(_ sdkTypes.StoreKey) sdkTypes.Store {
	panic("implement me")
}

func (t TestMultiStore) GetKVStore(key sdkTypes.StoreKey) sdkTypes.KVStore {
	if store, ok := t.kvstore[key.String()]; ok {
		return store
	} else {
		store := NewTestKVStore()
		t.kvstore[key.String()] = store
		return store
	}
}

func (t TestMultiStore) TracingEnabled() bool {
	panic("implement me")
}

func (t TestMultiStore) SetTracer(_ io.Writer) sdkTypes.MultiStore {
	panic("implement me")
}

func (t TestMultiStore) SetTracingContext(_ sdkTypes.TraceContext) sdkTypes.MultiStore {
	panic("implement me")
}

func NewTestKVStore() sdkTypes.KVStore {
	return TestKVStore{mutex: &sync.RWMutex{}, store: make(map[string][]byte)}
}

type TestKVStore struct {
	mutex *sync.RWMutex
	store map[string][]byte
}

func (t TestKVStore) GetStoreType() sdkTypes.StoreType {
	panic("implement me")
}

func (t TestKVStore) CacheWrap() sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestKVStore) CacheWrapWithTrace(_ io.Writer, _ sdkTypes.TraceContext) sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestKVStore) Get(key []byte) []byte {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	val, ok := t.store[string(key)]

	if ok {
		return val
	} else {
		return nil
	}
}

func (t TestKVStore) Has(key []byte) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	_, ok := t.store[string(key)]
	return ok
}

func (t TestKVStore) Set(key, value []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.store[string(key)] = value
}

func (t TestKVStore) Delete(key []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	delete(t.store, string(key))
}

func (t TestKVStore) Iterator(start, end []byte) sdkTypes.Iterator {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	return newMockIterator(start, end, t.store)
}

func (t TestKVStore) ReverseIterator(start, end []byte) sdkTypes.Iterator {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	mock := newMockIterator(start, end, t.store)

	// reverse the order of the iterator, which is returned already
	// sorted in ascending order
	for i, j := 0, len(mock.keys)-1; i < j; i, j = i+1, j-1 {
		mock.keys[i], mock.keys[j] = mock.keys[j], mock.keys[i]
		mock.values[i], mock.values[j] = mock.values[j], mock.values[i]

	}

	mock.start = end
	mock.end = start

	return mock
}

type TestStoreKey string

// NewKVStoreKey provides a simple store key for testing
func NewKVStoreKey(key string) sdkTypes.StoreKey {
	return TestStoreKey(key)
}

func (t TestStoreKey) Name() string {
	return string(t)
}

func (t TestStoreKey) String() string {
	return string(t)
}

// mock iterator
type mockIterator struct {
	keys       [][]byte
	values     [][]byte
	index      int
	start, end []byte
}

func newMockIterator(start, end []byte, content map[string][]byte) mockIterator {

	keys := make([][]byte, 0)

	// select the keys according to the specified domain
	for k := range content {

		b := []byte(k)

		if (start == nil && end == nil) || (bytes.Compare(b, start) >= 0 && bytes.Compare(b, end) < 0) {

			//make sure data is a copy so that there is no concurrent writing
			temp := make([]byte, len(k))
			copy(temp, []byte(k))
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

		//make sure data is a copy so that there is no concurrent writing
		value := content[string(keys[i])]
		temp := make([]byte, len(value))
		copy(temp, value)

		values[i] = temp
	}

	return mockIterator{
		keys:   keys,
		values: values,
		index:  0,
		start:  start,
		end:    end,
	}
}

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
func (mi mockIterator) Next() {

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
func (mi mockIterator) Close() {

	//Do what?
}
