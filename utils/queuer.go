package utils

import (
	"encoding/binary"
	"fmt"

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
	store   KVStore
	maxSize uint64
	seqNo   uint64
	size    uint64
	logger  log.Logger
}

var (
	sizeKey     = KeyFromStr("size")
	sequenceKey = KeyFromStr("sequence")
	queueKey    = KeyFromStr("queue")
)

// NewSequenceKVQueue is the constructor of SequenceKVQueue. The prefixStore should isolate the namespace of the queue
func NewSequenceKVQueue(prefixStore KVStore, maxSize uint64, logger log.Logger) SequenceKVQueue {
	var seqNo uint64
	bz := prefixStore.GetRaw(sequenceKey)
	if bz != nil {
		seqNo = binary.BigEndian.Uint64(bz)
	}

	var size uint64
	bz = prefixStore.GetRaw(sizeKey)
	if bz != nil {
		size = binary.BigEndian.Uint64(bz)
	}

	return SequenceKVQueue{store: prefixStore, maxSize: maxSize, seqNo: seqNo, size: size, logger: logger}
}

// Enqueue pushes the given value onto the top of the queue and stores the value at given key. Returns queue position and status
func (q *SequenceKVQueue) Enqueue(value codec.ProtoMarshaler) error {
	if q.size >= q.maxSize {
		return fmt.Errorf("queue is full")
	}

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, q.seqNo)

	q.store.Set(queueKey.Append(KeyFromBz(bz)), value)
	q.incrSize()
	q.incrSeqNo()

	return nil
}

// Dequeue pops the nth value off and unmarshals it into the given object, returns false if n is out of range
// n is 0-based
func (q *SequenceKVQueue) Dequeue(n uint64, value codec.ProtoMarshaler) bool {
	key, ok := q.peek(n, value)
	if ok {
		q.store.Delete(key)
		q.decrSize()
	}
	return ok
}

// Size returns the given current size of the queue
func (q SequenceKVQueue) Size() uint64 {
	return q.size
}

// Peek unmarshals the value into the given object at the given index and returns its sequence number, returns false if n is out of range
// n is 0-based
func (q SequenceKVQueue) Peek(n uint64, value codec.ProtoMarshaler) bool {
	_, ok := q.peek(n, value)
	return ok
}

func (q SequenceKVQueue) peek(n uint64, value codec.ProtoMarshaler) (Key, bool) {
	if n >= q.size {
		return nil, false
	}

	var i uint64
	iter := q.store.Iterator(queueKey)
	defer CloseLogError(iter, q.logger)

	for ; iter.Valid() && i < n; iter.Next() {
		i++
	}
	iter.UnmarshalValue(value)
	return iter.GetKey(), true
}

func (q *SequenceKVQueue) incrSize() {
	q.size++
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, q.size)
	q.store.SetRaw(sizeKey, bz)
}

func (q *SequenceKVQueue) decrSize() {
	if q.size > 0 {
		q.size--
		bz := make([]byte, 8)
		binary.BigEndian.PutUint64(bz, q.size)
		q.store.SetRaw(sizeKey, bz)
	}
}

func (q *SequenceKVQueue) incrSeqNo() {
	q.seqNo++
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, q.seqNo)
	q.store.SetRaw(sequenceKey, bz)
}
