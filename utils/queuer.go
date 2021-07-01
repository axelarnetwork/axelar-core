package utils

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
)

//go:generate moq -out ./mock/queuer.go -pkg mock . KVQueue

// KVQueue represents a queue built with the KVStore
type KVQueue interface {
	Enqueue(key Keyer, value codec.ProtoMarshaler)
	Dequeue(value codec.ProtoMarshaler) bool
}

// BlockHeightKVQueue is a queue that orders items with the block height at which the items are enqueued;
// the order of items that are enqueued at the same block height is deterministically based on their actual key
// in the KVStore
type BlockHeightKVQueue struct {
	store             NormalizedKVStore
	blockHeightPrefix string
	name              string
}

// NewBlockHeightKVQueue is the constructor of BlockHeightKVQueue
func NewBlockHeightKVQueue(name string, store NormalizedKVStore, blockHeight int64) BlockHeightKVQueue {
	return BlockHeightKVQueue{store: store, name: name}.WithBlockHeight(blockHeight)
}

// Enqueue pushes the given value onto the top of the queue and stores the value at given key
func (q BlockHeightKVQueue) Enqueue(key Keyer, value codec.ProtoMarshaler) {
	q.store.Set(key.WithPrefix(q.blockHeightPrefix).WithPrefix(q.name), &gogoprototypes.BytesValue{Value: key.AsKey()})
	q.store.Set(key, value)
}

// Dequeue pops the bottom of the queue and stores it at the given value, and return true if anything
// in the queue is found
func (q BlockHeightKVQueue) Dequeue(value codec.ProtoMarshaler) bool {
	iter := sdk.KVStorePrefixIterator(q.store.KVStore, []byte(q.name))
	//goland:noinspection GoUnhandledErrorResult
	defer iter.Close()

	if !iter.Valid() {
		return false
	}

	var key gogoprototypes.BytesValue
	q.store.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &key)

	if ok := q.store.Get(RegularKey(key.Value), value); !ok {
		return false
	}

	q.store.Delete(RegularKey(iter.Key()))

	return true
}

func (q BlockHeightKVQueue) WithBlockHeight(blockHeight int64) BlockHeightKVQueue {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(blockHeight))
	q.blockHeightPrefix = string(bz)
	return q
}
