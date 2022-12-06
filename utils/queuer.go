package utils

import (
	"encoding/binary"
	"fmt"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/utils/funcs"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
)

var (
	currIdx = key.FromUInt[uint](1)
	queue   = key.FromUInt[uint](2)
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

type Queue[T ValidatedProtoMarshaler] struct {
	name        key.Key
	store       KVStore
	prioritizer func(value T) key.Key
}

func NewQueue[T ValidatedProtoMarshaler](name key.Key, store KVStore, prioritizer func(value T) key.Key) Queue[T] {
	return Queue[T]{
		name:        name,
		store:       store,
		prioritizer: prioritizer,
	}
}

func (q Queue[T]) Enqueue(value T) error {
	idx := q.getCurrIdx()
	if err := q.store.SetNewValidated(q.name.
		Append(queue).
		Append(q.prioritizer(value)).
		Append(key.FromUInt(idx)), value); err != nil {
		return err
	}

	return q.setCurrIdx(idx + 1)
}

func (q Queue[T]) Dequeue(value ...T) bool {
	iter := q.store.IteratorNew(q.name.Append(queue))

	if !iter.Valid() {
		funcs.MustNoErr(iter.Close())
		return false
	}

	iKey := iter.GetKeyNew()
	funcs.MustNoErr(iter.Close())

	for _, v := range value {
		iter.UnmarshalValue(v)
	}

	q.store.DeleteNew(iKey)
	return true
}

func (q Queue[T]) Peek(value ...T) bool {
	iter := q.store.IteratorNew(q.name.Append(queue))
	defer funcs.MustNoErr(iter.Close())

	if !iter.Valid() {
		return false
	}

	for _, v := range value {
		iter.UnmarshalValue(v)
	}
	return true
}

func (q Queue[T]) List() QueueIterator[T] {
	return QueueIterator[T]{iter: q.store.IteratorNew(q.name.Append(queue))}
}

func (q Queue[T]) getCurrIdx() uint64 {
	idx := gogoprototypes.UInt64Value{}
	if !q.store.GetNew(q.name.Append(currIdx), &idx) {
		return 0
	}

	return idx.Value
}

func (q Queue[T]) setCurrIdx(idx uint64) error {
	return q.store.SetNewValidated(q.name.Append(currIdx), NoValidation(&gogoprototypes.UInt64Value{Value: idx}))
}

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

// Value returns the value at the current position. Panics if the iterator is invalid.
// CONTRACT: value readonly []byte
func (i QueueIterator[T]) Value(value T) {
	i.iter.UnmarshalValue(value)
}

func (i QueueIterator[T]) Close() error {
	return i.iter.Close()
}
