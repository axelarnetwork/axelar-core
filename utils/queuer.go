package utils

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
)

//go:generate moq -out ./mock/queuer.go -pkg mock . KVQueue

// KVQueue represents a queue built with the KVStore
type KVQueue interface {
	// Enqueue pushes the given value into the queue with the given key
	Enqueue(key Key, value codec.ProtoMarshaler)
	// Dequeue pops the first item in queue and stores it in the given value
	Dequeue(value codec.ProtoMarshaler) bool
	// DequeueIf pops the first item in queue iff it matches the given filter and stores it in the given value
	DequeueIf(value codec.ProtoMarshaler, filter func(value codec.ProtoMarshaler) bool) bool
	// DequeueUntil pops the first item in queue that matches the given filter and stores it in the given value
	DequeueUntil(value codec.ProtoMarshaler, filter func(value codec.ProtoMarshaler) bool) bool
	// IsEmpty returns true if the queue is empty; false otherwise
	IsEmpty() bool

	//TODO: convert to iterator
	Keys() []Key
}

var _ KVQueue = GeneralKVQueue{}
var _ KVQueue = BlockHeightKVQueue{}

// GeneralKVQueue is a queue that orders items based on the given prioritizer function
type GeneralKVQueue struct {
	name        StringKey
	store       KVStore
	logger      log.Logger
	prioritizer func(value codec.ProtoMarshaler) Key
}

// NewGeneralKVQueue is the contructor for GeneralKVQueue
func NewGeneralKVQueue(name string, store KVStore, logger log.Logger, prioritizer func(value codec.ProtoMarshaler) Key) GeneralKVQueue {
	return GeneralKVQueue{
		name:        KeyFromStr(name),
		store:       store,
		logger:      logger,
		prioritizer: prioritizer,
	}
}

// Enqueue pushes the given value onto the top of the queue and stores the value at given key
func (q GeneralKVQueue) Enqueue(key Key, value codec.ProtoMarshaler) {
	q.store.Set(q.name.Append(q.prioritizer(value)).Append(key), &gogoprototypes.BytesValue{Value: key.AsKey()})
	q.store.Set(key, value)
}

func noopFilter(_ codec.ProtoMarshaler) bool {
	return true
}

// Dequeue pops the first item in queue and stores it in the given value
func (q GeneralKVQueue) Dequeue(value codec.ProtoMarshaler) bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	return q.dequeue(value, iter, noopFilter, false)
}

// DequeueIf pops the first item in queue iff it matches the given filter and stores it in the given value
func (q GeneralKVQueue) DequeueIf(value codec.ProtoMarshaler, filter func(value codec.ProtoMarshaler) bool) bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	return q.dequeue(value, iter, filter, false)
}

// DequeueUntil pops the first item in queue that matches the given filter and stores it in the given value
func (q GeneralKVQueue) DequeueUntil(value codec.ProtoMarshaler, filter func(value codec.ProtoMarshaler) bool) bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	return q.dequeue(value, iter, filter, true)
}

func (q GeneralKVQueue) dequeue(value codec.ProtoMarshaler, iter db.Iterator, filter func(value codec.ProtoMarshaler) bool, continueIfNotQualified bool) bool {
	for ; iter.Valid(); iter.Next() {
		var key gogoprototypes.BytesValue
		q.store.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &key)

		if ok := q.store.Get(KeyFromBz(key.Value), value); !ok {
			return false
		}

		isQualified := filter(value)
		if isQualified {
			q.store.Delete(KeyFromBz(iter.Key()))

			return true
		}

		if !continueIfNotQualified {
			return false
		}
	}

	return false
}

// IsEmpty returns true if the queue is empty; otherwise, false
func (q GeneralKVQueue) IsEmpty() bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	return !iter.Valid()
}

// Keys returns a list with the keys for the values still enqueued
func (q GeneralKVQueue) Keys() []Key {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	var keys []Key
	for ; iter.Valid(); iter.Next() {
		var key gogoprototypes.BytesValue
		q.store.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &key)

		keys = append(keys, KeyFromBz(key.Value))
	}

	return keys
}

// ExportState exports the given queue's state from the kv store
func (q GeneralKVQueue) ExportState() (state QueueState) {
	state.Items = make(map[string]QueueState_Item)

	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	for ; iter.Valid(); iter.Next() {
		var key gogoprototypes.BytesValue
		q.store.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &key)

		item := QueueState_Item{
			Key:   key.Value,
			Value: q.store.GetRaw(KeyFromBz(key.Value)),
		}

		state.Items[string(iter.Key())] = item
	}

	return state
}

// ImportState imports the given queue state into the kv store
func (q GeneralKVQueue) ImportState(state QueueState) {
	name := string(q.name.AsKey())
	for key, item := range state.Items {
		if !strings.HasPrefix(key, fmt.Sprintf("%s%s", name, DefaultDelimiter)) {
			panic(fmt.Errorf("queue key %s is invalid for queue %s", key, name))
		}

		q.store.Set(KeyFromStr(key), &gogoprototypes.BytesValue{Value: item.Key})
		q.store.SetRaw(KeyFromBz(item.Key), item.Value)
	}
}

// ValidateBasic returns an error if the given queue state is invalid
func (m QueueState) ValidateBasic(queueName ...string) error {
	itemKeySeen := make(map[string]bool)
	for key, item := range m.Items {
		if itemKeySeen[string(item.Key)] {
			return fmt.Errorf("duplicate item key")
		}

		if len(key) == 0 {
			return fmt.Errorf("queue key cannot be empty")
		}

		if len(queueName) > 0 && !strings.HasPrefix(key, fmt.Sprintf("%s%s", queueName[0], DefaultDelimiter)) {
			return fmt.Errorf("queue key %s is invalid for queue %s", key, queueName[0])
		}

		if len(item.Key) == 0 {
			return fmt.Errorf("item key cannot be empty")
		}

		if len(item.Value) == 0 {
			return fmt.Errorf("item value cannot be empty")
		}

		itemKeySeen[string(item.Key)] = true
	}

	return nil
}

// BlockHeightKVQueue is a queue that orders items with the block height at which the items are enqueued;
// the order of items that are enqueued at the same block height is deterministically based on their actual key
// in the KVStore
type BlockHeightKVQueue struct {
	GeneralKVQueue
}

// NewBlockHeightKVQueue is the constructor of BlockHeightKVQueue
func NewBlockHeightKVQueue(name string, store KVStore, blockHeight int64, logger log.Logger) BlockHeightKVQueue {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(blockHeight))

	return BlockHeightKVQueue{
		GeneralKVQueue: NewGeneralKVQueue(name, store, logger, func(_ codec.ProtoMarshaler) Key {
			return KeyFromBz(bz)
		}),
	}
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
