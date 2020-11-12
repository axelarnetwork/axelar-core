package mock

import (
	"io"

	sdkTypes "github.com/cosmos/cosmos-sdk/types"
)

func NewMultiStore() sdkTypes.MultiStore {
	return TestMultiStore{kvstore: map[string]sdkTypes.KVStore{}}
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

func (t TestMultiStore) CacheWrapWithTrace(w io.Writer, tc sdkTypes.TraceContext) sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestMultiStore) CacheMultiStore() sdkTypes.CacheMultiStore {
	panic("implement me")
}

func (t TestMultiStore) CacheMultiStoreWithVersion(version int64) (sdkTypes.CacheMultiStore, error) {
	panic("implement me")
}

func (t TestMultiStore) GetStore(key sdkTypes.StoreKey) sdkTypes.Store {
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

func (t TestMultiStore) SetTracer(w io.Writer) sdkTypes.MultiStore {
	panic("implement me")
}

func (t TestMultiStore) SetTracingContext(context sdkTypes.TraceContext) sdkTypes.MultiStore {
	panic("implement me")
}

func NewTestKVStore() sdkTypes.KVStore {
	return TestKVStore{store: map[string][]byte{}}
}

type TestKVStore struct {
	store map[string][]byte
}

func (t TestKVStore) GetStoreType() sdkTypes.StoreType {
	panic("implement me")
}

func (t TestKVStore) CacheWrap() sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestKVStore) CacheWrapWithTrace(w io.Writer, tc sdkTypes.TraceContext) sdkTypes.CacheWrap {
	panic("implement me")
}

func (t TestKVStore) Get(key []byte) []byte {
	val, ok := t.store[string(key)]

	if ok {
		return val
	} else {
		return nil
	}
}

func (t TestKVStore) Has(key []byte) bool {
	_, ok := t.store[string(key)]
	return ok
}

func (t TestKVStore) Set(key, value []byte) {
	t.store[string(key)] = value
}

func (t TestKVStore) Delete(key []byte) {
	delete(t.store, string(key))
}

func (t TestKVStore) Iterator(start, end []byte) sdkTypes.Iterator {
	panic("implement me")
}

func (t TestKVStore) ReverseIterator(start, end []byte) sdkTypes.Iterator {
	panic("implement me")
}
