package utils

import (
	"encoding/binary"
	"strconv"
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
	. "github.com/axelarnetwork/utils/test"
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
	var (
		kvQueue   GeneralKVQueue
		itemCount int64
	)

	repeats := 20

	whenHavingASequentialKVQueue := When("a sequential kv queue", func() {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		itemCount = rand.I64Between(10, 1000)
		items := make([]gogoprototypes.UInt64Value, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = gogoprototypes.UInt64Value{Value: uint64(i)}
		}

		kvQueue = NewGeneralKVQueue(rand.Str(5), store, log.TestingLogger(), func(value codec.ProtoMarshaler) Key {
			v := value.(*gogoprototypes.UInt64Value)
			bz := make([]byte, 8)
			binary.BigEndian.PutUint64(bz, v.Value)

			return KeyFromBz(bz)
		})

		for _, item := range items {
			kvQueue.Enqueue(KeyFromStr(strconv.FormatUint(item.Value, 10)), &item)
		}
	})

	t.Run("Dequeue", testutils.Func(func(t *testing.T) {
		whenHavingASequentialKVQueue.
			Then("should dequeue first item in queue", func(t *testing.T) {
				var actual gogoprototypes.UInt64Value
				assert.True(t, kvQueue.Dequeue(&actual))
				assert.EqualValues(t, 0, actual.Value)
			}).
			Run(t)
	}).Repeat(repeats))

	t.Run("DequeueIf", testutils.Func(func(t *testing.T) {
		whenHavingASequentialKVQueue.
			Then("should dequeue nothing if first item in queue does not match filter", func(t *testing.T) {
				var actual gogoprototypes.UInt64Value
				assert.False(t, kvQueue.DequeueIf(&actual, func(value codec.ProtoMarshaler) bool {
					return value.(*gogoprototypes.UInt64Value).Value > 0
				}))
			}).
			Run(t)

		whenHavingASequentialKVQueue.
			Then("should dequeue first item in queue if it matches filter", func(t *testing.T) {
				var actual gogoprototypes.UInt64Value
				assert.True(t, kvQueue.DequeueIf(&actual, func(value codec.ProtoMarshaler) bool {
					return value.(*gogoprototypes.UInt64Value).Value <= 0
				}))
				assert.EqualValues(t, 0, actual.Value)
			}).
			Run(t)
	}).Repeat(repeats))

	t.Run("DequeueUntil", testutils.Func(func(t *testing.T) {
		whenHavingASequentialKVQueue.
			Then("should dequeue the first item that matches the filter", func(t *testing.T) {
				min := uint64(rand.I64Between(0, itemCount-1))
				var actual gogoprototypes.UInt64Value
				assert.True(t, kvQueue.DequeueUntil(&actual, func(value codec.ProtoMarshaler) bool {
					return value.(*gogoprototypes.UInt64Value).Value >= min
				}))
				assert.EqualValues(t, min, actual.Value)
			}).
			Run(t)

		whenHavingASequentialKVQueue.
			Then("should dequeue nothing if no item in queue matches filter", func(t *testing.T) {
				var actual gogoprototypes.UInt64Value
				assert.False(t, kvQueue.DequeueUntil(&actual, func(value codec.ProtoMarshaler) bool {
					return value.(*gogoprototypes.UInt64Value).Value >= uint64(itemCount)
				}))
			}).
			Run(t)
	}).Repeat(repeats))

	t.Run("Enqueue and Dequeue", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		itemCount := rand.I64Between(10, 100)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		blockHeight := rand.I64Between(1, 10000)
		kvQueue := NewBlockHeightKVQueue("test-enqueue-dequeue", store, blockHeight, log.TestingLogger())

		for _, item := range items {
			kvQueue.Enqueue(KeyFromStr(item), &gogoprototypes.StringValue{Value: item})
			blockHeight += rand.I64Between(1, 1000)
			kvQueue = NewBlockHeightKVQueue("test-enqueue-dequeue", store, blockHeight, log.TestingLogger())
		}

		var actualItems []string
		var actualItem gogoprototypes.StringValue
		for kvQueue.Dequeue(&actualItem) {
			actualItems = append(actualItems, actualItem.Value)
		}
		assert.Equal(t, items, actualItems)
	}).Repeat(repeats))

	t.Run("ImportState and ExportState", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		itemCount := rand.I64Between(10, 100)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		blockHeight := rand.I64Between(1, 10000)
		kvQueue := NewBlockHeightKVQueue("test-queue", store, blockHeight, log.TestingLogger())

		for _, item := range items {
			kvQueue.Enqueue(KeyFromStr(item), &gogoprototypes.StringValue{Value: item})
			blockHeight += rand.I64Between(1, 1000)
			kvQueue = NewBlockHeightKVQueue("test-queue", store, blockHeight, log.TestingLogger())
		}

		state := kvQueue.ExportState()

		var expected []string
		var item gogoprototypes.StringValue
		for kvQueue.Dequeue(&item) {
			expected = append(expected, item.Value)
		}

		kvQueue.ImportState(state)
		var actual []string
		for kvQueue.Dequeue(&item) {
			actual = append(actual, item.Value)
		}

		assert.Equal(t, expected, actual)
	}).Repeat(repeats))
}

func TestNewSequenceKVQueue(t *testing.T) {
	repeats := 20

	t.Run("enqueue within limit and dequeue", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()

		queueSize := rand.I64Between(10, 100)
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

		queueSize := rand.I64Between(10, 100)
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

		queueSize := rand.I64Between(10, 100)
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

		queueSize := rand.I64Between(2, 1000)
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
