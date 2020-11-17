package mock

import (
	"io"
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

func (t TestKVStore) Iterator(_, _ []byte) sdkTypes.Iterator {
	panic("implement me")
}

func (t TestKVStore) ReverseIterator(_, _ []byte) sdkTypes.Iterator {
	panic("implement me")
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
