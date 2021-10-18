package utils

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
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
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		queueSize := rand.I64Between(10, 1000)
		itemCount := rand.I64Between(0, queueSize)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		currSize := int64(0)
		kvQueue := NewSequenceKVQueue("test-enqueue-dequeue", store, queueSize, log.TestingLogger())
		for _, item := range items {
			pos, err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: item})
			currSize++
			assert.Equal(t, err, nil)
			assert.Equal(t, pos, currSize)
			assert.Equal(t, kvQueue.Size(), currSize)
		}

		var actualItems []string
		var actualItem gogoprototypes.StringValue
		for seq := int64(0); seq < itemCount; seq++ {
			kvQueue.DequeueSequence(&actualItem, seq)
			actualItems = append(actualItems, actualItem.Value)
		}
		assert.Equal(t, items, actualItems)
	}).Repeat(repeats))

	t.Run("peek ith item, where i is smaller than queue size", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		queueSize := rand.I64Between(10, 1000)
		itemCount := rand.I64Between(1, queueSize)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var currSize int64
		kvQueue := NewSequenceKVQueue("test-enqueue-dequeue", store, queueSize, log.TestingLogger())
		for _, item := range items {
			pos, err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: item})
			currSize++
			assert.Equal(t, err, nil)
			assert.Equal(t, pos, currSize)
			assert.Equal(t, kvQueue.Size(), currSize)
		}

		for i := int64(0); i < itemCount; i++ {
			var actualItem gogoprototypes.StringValue
			seq := kvQueue.Peek(i, &actualItem)
			assert.Equal(t, items[i], actualItem.Value)
			assert.Equal(t, i, seq)
		}

	}).Repeat(repeats))

	t.Run("dequeue the last item", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		queueSize := rand.I64Between(10, 1000)
		itemCount := rand.I64Between(1, queueSize)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var currSize int64
		kvQueue := NewSequenceKVQueue("test-enqueue-dequeue", store, queueSize, log.TestingLogger())
		for _, item := range items {
			pos, err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: item})
			currSize ++
			assert.Equal(t, err, nil)
			assert.Equal(t, pos, currSize)
			assert.Equal(t, kvQueue.Size(), currSize)
		}

		var actualItems []string
		var actualItem gogoprototypes.StringValue
		var reverseItems []string
		for i := itemCount - 1; i > 0; i-- {
			seq := kvQueue.Peek(i, &actualItem)
			_ = kvQueue.DequeueSequence(&actualItem, seq)
			actualItems = append(actualItems, actualItem.Value)
			reverseItems = append(reverseItems, items[i])
		}

		assert.Equal(t, reverseItems, actualItems)

	}).Repeat(repeats))

	t.Run("should return error when enqueue item when queue size is full", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		queueSize := rand.I64Between(0, 1000)
		itemCount := queueSize + 1
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var currSize int64
		kvQueue := NewSequenceKVQueue("test-enqueue-dequeue", store, queueSize, log.TestingLogger())
		var i int64
		for ; i < itemCount-1; i++ {
			pos, err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: items[i]})
			currSize++
			assert.Equal(t, err, nil)
			assert.Equal(t, pos, currSize)
			assert.Equal(t, kvQueue.Size(), currSize)
		}
		pos, err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: items[i]})
		assert.Error(t, err)
		assert.Equal(t, pos, int64(-1))
		assert.Equal(t, kvQueue.Size(), currSize)

	}).Repeat(repeats))

	t.Run("should return false when dequeue sequence item when sequence is not found", testutils.Func(func(t *testing.T) {
		ctx, cdc := setup()
		store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		queueSize := rand.I64Between(0, 1000)
		itemCount := rand.I64Between(1, queueSize)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		var currSize int64
		kvQueue := NewSequenceKVQueue("test-enqueue-dequeue", store, queueSize, log.TestingLogger())
		var i int64
		for ; i < itemCount-1; i++ {
			pos, err := kvQueue.Enqueue(&gogoprototypes.StringValue{Value: items[i]})
			currSize++
			assert.Equal(t, err, nil)
			assert.Equal(t, pos, currSize)
			assert.Equal(t, kvQueue.Size(), currSize)
		}

		var actualItem gogoprototypes.StringValue

		found := kvQueue.DequeueSequence(&actualItem, itemCount)
		assert.Equal(t, false, found)

	}).Repeat(repeats))

}
