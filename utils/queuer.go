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

	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	currIdx = key.FromUInt[uint](1)
	queue   = key.FromUInt[uint](2)
	index   = key.FromUInt[uint](3)
)

// GeneralKVQueue is a queue that orders items based on the given prioritizer function
// Deprecated
type GeneralKVQueue struct {
	name        StringKey
	store       KVStore
	logger      log.Logger
	prioritizer func(value codec.ProtoMarshaler) Key
}

// NewGeneralKVQueue is the constructor for GeneralKVQueue
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
	for itemKey, item := range m.Items {
		if itemKeySeen[string(item.Key)] {
			return fmt.Errorf("duplicate item key")
		}

		if len(itemKey) == 0 {
			return fmt.Errorf("queue key cannot be empty")
		}

		// TODO: migrate for new queue
		if len(queueName) > 0 && !strings.HasPrefix(itemKey, fmt.Sprintf("%s%s", queueName[0], DefaultDelimiter)) {
			return fmt.Errorf("queue key %s is invalid for queue %s", itemKey, queueName[0])
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

// Queue is a generic persistent queue which is able to add priority to enqueued items
type Queue[T ValidatedProtoMarshaler] struct {
	store       KVStore
	prioritizer func(value T) key.Key
}

// NewQueue creates a new queue persisting items to the given name namespace. Use the prioritizer function to add a prefix to the item to adjust its queue ranking
func NewQueue[T ValidatedProtoMarshaler](store KVStore, name key.Key, prioritizer func(value T) key.Key) Queue[T] {
	return Queue[T]{
		store:       NewPrefixStore(store, name),
		prioritizer: prioritizer,
	}
}

// Enqueue adds an item to the queue with a ranking based on the prioritizer. Items with the same priority queue get sorted in FIFO order.
// Returns an error if the item's ValidateBasic() function returns an error
func (q Queue[T]) Enqueue(value T) error {
	_, err := q.enqueue(value)
	return err
}

// Dequeue pops an item from the front of the queue and unmarshals it into the given value. Returns true if the queue is not empty
func (q Queue[T]) Dequeue(value T) bool {
	qKey, ok := q.peek(value)
	if !ok {
		return false
	}

	q.store.DeleteNew(qKey)
	return true
}

// Discard pops an item from the front of the queue and discards it. Returns true if the queue is not empty
func (q Queue[T]) Discard() bool {
	qKey, ok := q.any()
	if !ok {
		return false
	}

	q.store.DeleteNew(qKey)
	return true
}

// Peek looks up the item at the front of the queue and unmarshals it into the given value, but leaves the queue as is.
// Returns true if the queue is not empty
func (q Queue[T]) Peek(value T) bool {
	_, ok := q.peek(value)
	return ok
}

// Any returns true if the queue is not empty
func (q Queue[T]) Any() bool {
	_, ok := q.any()

	return ok
}

// Iter returns an iterator over all elements of the queue according to the queue ranking. This is intended to be a read-only operation.
// CONTRACT: Close the iterator befor modifying the underlying data, e.g. enqueueing/dequeueing items
func (q Queue[T]) Iter() QueueIterator[T] {
	return QueueIterator[T]{iter: q.store.IteratorNew(queue)}
}

func (q Queue[T]) enqueue(value T) (key.Key, error) {
	idx := q.getCurrIdx()
	qKey := queue.
		Append(q.prioritizer(value)).
		Append(key.FromUInt(idx))
	if err := q.store.SetNewValidated(qKey, value); err != nil {
		return nil, err
	}

	return qKey, q.setCurrIdx(idx + 1)
}

func (q Queue[T]) any() (key.Key, bool) {
	iter := q.Iter().iter
	defer funcs.MustNoErr(iter.Close())

	if !iter.Valid() {
		return nil, false
	}

	return iter.GetKeyNew(), true
}

func (q Queue[T]) peek(value codec.ProtoMarshaler) (key.Key, bool) {
	iter := q.Iter().iter
	defer funcs.MustNoErr(iter.Close())

	if !iter.Valid() {
		return nil, false
	}

	iter.UnmarshalValue(value)
	iKey := iter.GetKeyNew()

	return iKey, true
}

func (q Queue[T]) getCurrIdx() uint64 {
	idx := gogoprototypes.UInt64Value{}
	if !q.store.GetNew(currIdx, &idx) {
		return 0
	}

	return idx.Value
}

func (q Queue[T]) setCurrIdx(idx uint64) error {
	return q.store.SetNewValidated(currIdx, NoValidation(&gogoprototypes.UInt64Value{Value: idx}))
}

// QueueIterator is an iterator over elements of a queue
type QueueIterator[T ValidatedProtoMarshaler] struct {
	iter Iterator
}

// Valid returns whether the current iterator is valid. Once invalid, the Iterator remains
// invalid forever.
func (i QueueIterator[T]) Valid() bool {
	return i.iter.Valid()
}

// Next moves the iterator to the next key in the database, as defined by order of iteration.
// If Valid returns false, this method will panic.
func (i QueueIterator[T]) Next() {
	i.iter.Next()
}

// Value unamrshals the element at the current position into the value. Panics if the iterator is invalid.
func (i QueueIterator[T]) Value(value T) {
	i.iter.UnmarshalValue(value)
}

// Close cleans up the iterator, potentially releasing locks on the underlying datastore
func (i QueueIterator[T]) Close() error {
	return i.iter.Close()
}

// IndexedQueue is a queue that also maintains a lookup index for elements in the queue
type IndexedQueue[T ValidatedProtoMarshaler, S any] struct {
	q       Queue[T]
	indexer func(value T) S
}

// NewIndexedQueue returns a new queue with index according to the indexer function
// Example:
//
//	type element struct{
//	    ID int
//	}
//
// NewIndexedQueue(queue, func(e element) int { return e.ID }
func NewIndexedQueue[T ValidatedProtoMarshaler, S any](q Queue[T], indexer func(value T) S) IndexedQueue[T, S] {
	return IndexedQueue[T, S]{
		q:       q,
		indexer: indexer,
	}
}

// Enqueue adds an item to the end of the queue and indexes it
func (iq IndexedQueue[T, S]) Enqueue(value T) error {
	qKey, err := iq.q.enqueue(value)
	if err != nil {
		return err
	}
	return iq.index(value, qKey)
}

// Dequeue pops an item from the front of the queue and deletes its index
func (iq IndexedQueue[T, S]) Dequeue(value T) bool {
	iq.q.Dequeue(value)
	iq.deleteIndex(value)

	return true
}

// Has returns true if there is an element in the queue that is indexed by the given lookup value
func (iq IndexedQueue[T, S]) Has(lookup S) bool {
	return iq.q.store.HasNew(index.Append(key.FromAny(lookup)))
}

// Get returns true if there is an element in the queue that is indexed by the given lookup object and unmarshals the item into the value object
func (iq IndexedQueue[T, S]) Get(lookup S, value T) bool {
	var qKey gogoprototypes.BytesValue
	ok := iq.q.store.GetNew(index.Append(key.FromAny(lookup)), &qKey)

	if ok {
		iq.q.store.GetNew(key.FromRaw(qKey.Value), value)
	}

	return ok
}

func (iq IndexedQueue[T, S]) index(value T, qKey key.Key) error {
	return iq.q.store.SetNewValidated(index.Append(key.FromAny(iq.indexer(value))), NoValidation(&gogoprototypes.BytesValue{Value: qKey.Bytes()}))
}

func (iq IndexedQueue[T, S]) deleteIndex(value T) {
	iq.q.store.DeleteNew(index.Append(key.FromAny(iq.indexer(value))))
}
