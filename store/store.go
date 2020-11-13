package store

import (
	"sync"
)

/*
This store should be used to keep subjective local validator data.
The read and write access to this store must be tread-safe because validators might run
other processes (like broadcasting) concurrently to the main execution thread and all processes need access to the store.
*/
type SubjectiveStore struct {
	mutex *sync.RWMutex
	store map[string][]byte
}

func NewSubjectiveStore() SubjectiveStore {
	return SubjectiveStore{
		mutex: &sync.RWMutex{},
		store: make(map[string][]byte),
	}
}

// Get returns nil iff key doesn't exist. Panics on nil key.
func (s SubjectiveStore) Get(key []byte) []byte {
	if key == nil {
		panic("key cannot be nil")
	}
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	val, ok := s.store[string(key)]
	if !ok {
		return nil
	}
	return val
}

// Has checks if a key exists. Panics on nil key.
func (s SubjectiveStore) Has(key []byte) bool {
	if key == nil {
		panic("key cannot be nil")
	}
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, ok := s.store[string(key)]
	return ok
}

// Set sets the key. Panics on nil key or value.
func (s SubjectiveStore) Set(key, value []byte) {
	if key == nil {
		panic("key cannot be nil")
	}
	if value == nil {
		panic("value cannot be nil")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[string(key)] = value
}

// Delete deletes the key. Panics on nil key.
func (s SubjectiveStore) Delete(key []byte) {
	if key == nil {
		panic("key cannot be nil")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, string(key))
}
