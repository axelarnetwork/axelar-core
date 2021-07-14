package utils

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"
)

//go:generate moq -out ./mock/queuer.go -pkg mock . KVQueue

// KVQueue represents a queue built with the KVStore
type KVQueue interface {
	Enqueue(key Key, value codec.ProtoMarshaler)
	Dequeue(value codec.ProtoMarshaler) bool
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

// Dequeue pops the bottom of the queue and stores it at the given value, and return true if anything
// in the queue is found
func (q BlockHeightKVQueue) Dequeue(value codec.ProtoMarshaler) bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, q.name.AsKey())
	defer CloseLogError(iter, q.logger)

	if !iter.Valid() {
		return false
	}

	var key gogoprototypes.BytesValue
	q.store.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &key)

	if ok := q.store.Get(KeyFromBz(key.Value), value); !ok {
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
