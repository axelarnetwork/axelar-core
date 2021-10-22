package utils

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

var stringGen = rand.Strings(5, 50).Distinct()

func setup() (sdk.Context, *codec.ProtoCodec) {
	interfaceRegistry := types.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&gogoprototypes.StringValue{},
	)
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	return ctx, marshaler
}

func TestNewBlockHeightKVQueue(t *testing.T) {
	repeats := 20

	t.Run("enqueue and dequeue", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		itemCount := rand.I64Between(10, 1000)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		blockHeight := rand.I64Between(1, 10000)
		kvQueue := NewBlockHeightKVQueue("test-enqueue-dequeue", store, blockHeight, log.TestingLogger())
		for _, item := range items {
			kvQueue.Enqueue(KeyFromStr(item), &gogoprototypes.StringValue{Value: item})
			blockHeight += rand.I64Between(1, 1000)
			kvQueue = kvQueue.WithBlockHeight(blockHeight)
		}

		var actualItems []string
		var actualItem gogoprototypes.StringValue
		for kvQueue.Dequeue(&actualItem) {
			actualItems = append(actualItems, actualItem.Value)
		}
		assert.Equal(t, items, actualItems)
	}).Repeat(repeats))
}

func TestNewSequenceKVQueue(t *testing.T) {
	repeats := 20

	t.Run("enqueue within limit and dequeue", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()

		queueSize := rand.I64Between(10, 1000)
		itemCount := uint64(rand.I64Between(1, queueSize))
		items := make([]string, itemCount)

		store := prefix.NewStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), []byte("test-enqueue-dequeue"))
		kvQueue := NewSequenceKVQueue(NewNormalizedStore(store, cdc), uint64(queueSize), log.TestingLogger())
		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		i := uint64(0)
		for _, item := range items {
			err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: item})
			i++
			assert.Equal(t, err, nil)
			assert.Equal(t, i, kvQueue.Size())
		}

		var actualItems []string
		var actualItem gogoprototypes.StringValue
		for kvQueue.Dequeue(0, &actualItem) {
			actualItems = append(actualItems, actualItem.Value)
		}
		assert.Equal(t, items, actualItems)
	}).Repeat(repeats))

	t.Run("peek ith item, where i is smaller than queue size", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()

		queueSize := rand.I64Between(10, 1000)
		itemCount := rand.I64Between(1, queueSize)
		items := make([]string, itemCount)

		store := prefix.NewStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), []byte("test-enqueue-dequeue"))
		kvQueue := NewSequenceKVQueue(NewNormalizedStore(store, cdc), uint64(queueSize), log.TestingLogger())

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var i uint64
		for _, item := range items {
			err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: item})
			i++
			assert.Equal(t, err, nil)
			assert.Equal(t, kvQueue.Size(), i)
		}

		for idx := uint64(0); idx < uint64(itemCount); idx++ {
			var actualItem gogoprototypes.StringValue
			ok := kvQueue.Peek(idx, &actualItem)
			assert.Equal(t, items[idx], actualItem.Value)
			assert.Equal(t, true, ok)
		}

	}).Repeat(repeats))

	t.Run("dequeue the last item", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()

		queueSize := rand.I64Between(10, 1000)
		itemCount := rand.I64Between(1, queueSize)
		items := make([]string, itemCount)

		store := prefix.NewStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), []byte("test-enqueue-dequeue"))
		kvQueue := NewSequenceKVQueue(NewNormalizedStore(store, cdc), uint64(queueSize), log.TestingLogger())

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var i uint64
		for _, item := range items {
			err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: item})
			i++
			assert.Equal(t, err, nil)
			assert.Equal(t, kvQueue.Size(), i)
		}

		var actualItems []string
		var actualItem gogoprototypes.StringValue
		var reverseItems []string
		for idx := uint64(itemCount) - 1; idx > 0; idx-- {
			ok := kvQueue.Peek(idx, &actualItem)
			assert.Equal(t, true, ok)
			ok = kvQueue.Dequeue(idx, &actualItem)
			assert.Equal(t, true, ok)
			assert.Equal(t, kvQueue.Size(), idx)
			actualItems = append(actualItems, actualItem.Value)
			reverseItems = append(reverseItems, items[idx])
		}

		assert.Equal(t, reverseItems, actualItems)

	}).Repeat(repeats))

	t.Run("should return error when enqueue item when queue size is full", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()

		queueSize := rand.I64Between(0, 1000)
		itemCount := queueSize + 1
		items := make([]string, itemCount)

		store := prefix.NewStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), []byte("test-enqueue-dequeue"))
		kvQueue := NewSequenceKVQueue(NewNormalizedStore(store, cdc), uint64(queueSize), log.TestingLogger())

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var i uint64
		for ; i < uint64(itemCount)-1; i++ {
			err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: items[i]})
			assert.Equal(t, err, nil)
			assert.Equal(t, kvQueue.Size(), i+1)
		}
		err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: items[i]})
		assert.Error(t, err)
		assert.Equal(t, kvQueue.Size(), i)

	}).Repeat(repeats))

	t.Run("should return false when dequeue idx is out of index", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()

		queueSize := rand.I64Between(0, 1000)
		itemCount := uint64(rand.I64Between(1, queueSize))
		items := make([]string, itemCount)

		store := prefix.NewStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), []byte("test-enqueue-dequeue"))
		kvQueue := NewSequenceKVQueue(NewNormalizedStore(store, cdc), uint64(queueSize), log.TestingLogger())

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var i uint64
		for ; i < itemCount-1; i++ {
			err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: items[i]})
			assert.Equal(t, err, nil)
			assert.Equal(t, kvQueue.Size(), i+1)
		}

		var actualItem gogoprototypes.StringValue

		ok := kvQueue.Dequeue(itemCount, &actualItem)
		assert.Equal(t, false, ok)

	}).Repeat(repeats))

}
