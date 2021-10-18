package utils

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"
)

//go:generate moq -out ./mock/queuer.go -pkg mock . KVQueue

// KVQueue represents a queue built with the KVStore
type KVQueue interface {
	Enqueue(key Key, value codec.ProtoMarshaler)
	Dequeue(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool
	IsEmpty() bool
}

// SequenceQueue represents a queue built with the SequenceKVQueue
type SequenceQueue interface {
	Enqueue(value codec.ProtoMarshaler) (int64, error)
	DequeueSequence(value codec.ProtoMarshaler, sequence int64) bool
	Peek(n int64, value codec.ProtoMarshaler) int64
	Size() int64
}

// BlockHeightKVQueue is a queue that orders items with the block height at which the items are enqueued;
// the order of items that are enqueued at the same block height is deterministically based on their actual key
// in the KVStore
type BlockHeightKVQueue struct {
	store       KVStore
	blockHeight Key
	name        Key
	logger      log.Logger
}

// NewBlockHeightKVQueue is the constructor of BlockHeightKVQueue
func NewBlockHeightKVQueue(name string, store KVStore, blockHeight int64, logger log.Logger) BlockHeightKVQueue {
	return BlockHeightKVQueue{store: store, name: KeyFromStr(name), logger: logger}.WithBlockHeight(blockHeight)
}

// Enqueue pushes the given value onto the top of the queue and stores the value at given key
func (q BlockHeightKVQueue) Enqueue(key Key, value codec.ProtoMarshaler) {
	q.store.Set(q.name.Append(q.blockHeight).Append(key), &gogoprototypes.BytesValue{Value: key.AsKey()})
	q.store.Set(key, value)
}

// Dequeue pops the bottom of the queue and unmarshals it into the given object, and return true if anything
// in the queue is found and the value passes the optional filter function
func (q BlockHeightKVQueue) Dequeue(value codec.ProtoMarshaler, filter ...func(value codec.ProtoMarshaler) bool) bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	if !iter.Valid() {
		return false
	}

	var key gogoprototypes.BytesValue
	q.store.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &key)

	if ok := q.store.Get(KeyFromBz(key.Value), value); !ok {
		return false
	}

	if len(filter) > 0 && !filter[0](value) {
		return false
	}

	q.store.Delete(KeyFromBz(iter.Key()))

	return true
}

// IsEmpty returns true if the queue is empty; otherwise, false
func (q BlockHeightKVQueue) IsEmpty() bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	//goland:noinspection GoUnhandledErrorResult
	defer iter.Close()

	return !iter.Valid()
}

// WithBlockHeight returns a queue with the given block height
func (q BlockHeightKVQueue) WithBlockHeight(blockHeight int64) BlockHeightKVQueue {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(blockHeight))
	q.blockHeight = KeyFromBz(bz)
	return q
}

// SequenceKVQueue is a queue that orders items with the sequence number at which the items are enqueued;
type SequenceKVQueue struct {
	store    KVStore
	size     int64
	name     Key
	sequence Key
	logger   log.Logger
}

var (
	sizeKey     = KeyFromStr("size")
	sequenceKey = KeyFromStr("sequence")
)

// NewSequenceKVQueue is the constructor of SequenceKVQueue
func NewSequenceKVQueue(name string, store KVStore, size int64, logger log.Logger) SequenceKVQueue {
	return SequenceKVQueue{store: store, size: size, name: KeyFromStr(name), logger: logger}
}

// Enqueue pushes the given value onto the top of the queue and stores the value at given key. Returns queue position and status
func (q SequenceKVQueue) Enqueue(value codec.ProtoMarshaler) (int64, error) {
	if q.Size() >= q.size {
		return -1, fmt.Errorf("sign queue is full")
	}

	sequence := q.Sequence()
	sequenceBz := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceBz, uint64(sequence))

	q.store.Set(q.name.Append(KeyFromBz(sequenceBz)), value)
	q.increSize()
	q.increSequence()

	return q.Size(), nil
}

// DequeueSequence pops the value with the given sequence of the queue and unmarshals it into the given object, and return true if value if found
func (q SequenceKVQueue) DequeueSequence(value codec.ProtoMarshaler, sequence int64) bool {
	iter := q.store.Iterator(q.name)
	defer CloseLogError(iter, q.logger)

	sequenceBz := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceBz, uint64(sequence))

	for ; iter.Valid(); iter.Next() {
		if iter.GetKey().Equals(q.name.Append(KeyFromBz(sequenceBz))) {
			iter.UnmarshalValue(value)
			q.store.Delete(iter.GetKey())
			q.decreSize()
			return true
		}
	}
	return false
}

// Size returns the given current size of the queue
func (q SequenceKVQueue) Size() int64 {
	bz := q.store.GetRaw(sizeKey.Append(q.name))
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

// Sequence returns the current sequence of the queue
func (q SequenceKVQueue) Sequence() int64 {
	bz := q.store.GetRaw(sequenceKey.Append(q.name))
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

// Peek unmarshals the value into the given object at the given index and returns its sequence number.
// Returns -1 if index is out of range
func (q SequenceKVQueue) Peek(i int64, value codec.ProtoMarshaler) int64 {
	if i >= q.Size() {
		return -1
	}

	iter := q.store.Iterator(q.name)
	defer CloseLogError(iter, q.logger)

	var counter int64
	for ; iter.Valid() && counter < i; iter.Next() {
		counter++
	}

	iter.UnmarshalValue(value)
	seq := strings.TrimPrefix(string(iter.Key()), string(q.name.AsKey())+"_")
	return int64(binary.BigEndian.Uint64([]byte(seq)))
}

func (q SequenceKVQueue) increSize() {
	currSize := q.Size()
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(currSize)+1)
	q.store.SetRaw(sizeKey.Append(q.name), bz)
}

func (q SequenceKVQueue) decreSize() {
	currSize := q.Size()
	if currSize > 0 {
		bz := make([]byte, 8)
		binary.BigEndian.PutUint64(bz, uint64(currSize)-1)
		q.store.SetRaw(sizeKey.Append(q.name), bz)
	}
}

func (q SequenceKVQueue) increSequence() {
	currSequence := q.Sequence()
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(currSequence)+1)
	q.store.SetRaw(sequenceKey.Append(q.name), bz)
}
