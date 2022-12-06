package utils_test

import (
	"encoding/binary"
	"errors"
	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"strconv"
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
	. "github.com/axelarnetwork/utils/test"
)

var stringGen = rand.Strings(5, 50).Distinct()

func queue_setup() (sdk.Context, *codec.ProtoCodec) {
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
		kvQueue   utils.GeneralKVQueue
		itemCount int64
	)

	repeats := 20

	whenHavingASequentialKVQueue := When("a sequential kv queue", func() {
		ctx, cdc := queue_setup()
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		itemCount = rand.I64Between(10, 1000)
		items := make([]gogoprototypes.UInt64Value, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = gogoprototypes.UInt64Value{Value: uint64(i)}
		}

		kvQueue = utils.NewGeneralKVQueue(rand.Str(5), store, log.TestingLogger(), func(value codec.ProtoMarshaler) utils.Key {
			v := value.(*gogoprototypes.UInt64Value)
			bz := make([]byte, 8)
			binary.BigEndian.PutUint64(bz, v.Value)

			return utils.KeyFromBz(bz)
		})

		for _, item := range items {
			kvQueue.Enqueue(utils.KeyFromStr(strconv.FormatUint(item.Value, 10)), &item)
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
		ctx, cdc := queue_setup()
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		itemCount := rand.I64Between(10, 100)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		blockHeight := rand.I64Between(1, 10000)
		kvQueue := utils.NewBlockHeightKVQueue("test-enqueue-dequeue", store, blockHeight, log.TestingLogger())

		for _, item := range items {
			kvQueue.Enqueue(utils.KeyFromStr(item), &gogoprototypes.StringValue{Value: item})
			blockHeight += rand.I64Between(1, 1000)
			kvQueue = utils.NewBlockHeightKVQueue("test-enqueue-dequeue", store, blockHeight, log.TestingLogger())
		}

		var actualItems []string
		var actualItem gogoprototypes.StringValue
		for kvQueue.Dequeue(&actualItem) {
			actualItems = append(actualItems, actualItem.Value)
		}
		assert.Equal(t, items, actualItems)
	}).Repeat(repeats))

	t.Run("ImportState and ExportState", testutils.Func(func(t *testing.T) {
		ctx, cdc := queue_setup()
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey(stringGen.Next())), cdc)

		itemCount := rand.I64Between(10, 100)
		items := make([]string, itemCount)

		for i := 0; i < int(itemCount); i++ {
			items[i] = rand.Str(10)
		}

		blockHeight := rand.I64Between(1, 10000)
		kvQueue := utils.NewBlockHeightKVQueue("test-queue", store, blockHeight, log.TestingLogger())

		for _, item := range items {
			kvQueue.Enqueue(utils.KeyFromStr(item), &gogoprototypes.StringValue{Value: item})
			blockHeight += rand.I64Between(1, 1000)
			kvQueue = utils.NewBlockHeightKVQueue("test-queue", store, blockHeight, log.TestingLogger())
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

func TestQueue(t *testing.T) {
	var (
		queue utils.Queue[utils.ValidatedProtoMarshaler]
	)

	Given("a queue without prioritization", func() {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		cdc := app.MakeEncodingConfig().Codec
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("storeKey")), cdc)
		queue = utils.NewQueue(key.FromStr("queue1"), store, func(_ utils.ValidatedProtoMarshaler) key.Key { return key.FromBz([]byte{}) })
	}).
		When("enqueuing items", func() {
			assert.NoError(t, queue.Enqueue(utils.NoValidation(&gogoprototypes.UInt64Value{Value: 9})))
			assert.NoError(t, queue.Enqueue(utils.NoValidation(&gogoprototypes.UInt64Value{Value: 8})))
			assert.Error(t, queue.Enqueue(utils.WithValidation(&gogoprototypes.UInt64Value{Value: 6}, func() error {
				return errors.New("wrong input")
			})))
			assert.NoError(t, queue.Enqueue(utils.NoValidation(&gogoprototypes.UInt64Value{Value: 5})))
		}).
		Then("we can iterate over the queue", func(t *testing.T) {
			iter := queue.List()
			defer func() { _ = iter.Close() }()

			value := &gogoprototypes.UInt64Value{}
			validatedValue := utils.NoValidation(value)

			assert.True(t, iter.Valid())
			iter.Value(validatedValue)
			assert.EqualValues(t, 9, value.Value)
			iter.Next()

			assert.True(t, iter.Valid())
			iter.Value(validatedValue)
			assert.EqualValues(t, 8, value.Value)
			iter.Next()

			assert.True(t, iter.Valid())
			iter.Value(validatedValue)
			assert.EqualValues(t, 5, value.Value)
			iter.Next()

			assert.False(t, iter.Valid())
		}).
		Then("dequeue in the order they where enqueued", func(t *testing.T) {
			value := &gogoprototypes.UInt64Value{}
			validatedValue := utils.NoValidation(value)
			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 9, value.Value)
			assert.True(t, queue.Dequeue(validatedValue))
			assert.EqualValues(t, 9, value.Value)

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 8, value.Value)
			assert.True(t, queue.Dequeue(validatedValue))
			assert.EqualValues(t, 8, value.Value)

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 5, value.Value)
			assert.True(t, queue.Dequeue(validatedValue))
			assert.EqualValues(t, 5, value.Value)

			assert.False(t, queue.Peek(validatedValue))
			assert.False(t, queue.Dequeue(validatedValue))
		}).
		Then("iterator is empty after dequeue", func(t *testing.T) {
			iter := queue.List()
			defer func() { _ = iter.Close() }()

			assert.False(t, iter.Valid())
		}).Run(t)

	Given("a queue with prioritization", func() {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		cdc := app.MakeEncodingConfig().Codec
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("storeKey")), cdc)

		prio := uint(10)
		queue = utils.NewQueue(key.FromStr("queue1"), store, func(v utils.ValidatedProtoMarshaler) key.Key {
			defer func() { prio-- }()
			return key.FromUInt(prio)
		})
	}).
		When("enqueuing items", func() {
			assert.NoError(t, queue.Enqueue(utils.NoValidation(&gogoprototypes.UInt64Value{Value: 9})))
			assert.NoError(t, queue.Enqueue(utils.NoValidation(&gogoprototypes.UInt64Value{Value: 8})))
			assert.Error(t, queue.Enqueue(utils.WithValidation(&gogoprototypes.UInt64Value{Value: 6}, func() error {
				return errors.New("wrong input")
			})))
			assert.NoError(t, queue.Enqueue(utils.NoValidation(&gogoprototypes.UInt64Value{Value: 5})))
		}).
		Then("dequeue in the order they where enqueued", func(t *testing.T) {

			var value gogoprototypes.UInt64Value
			validatedValue := utils.NoValidation(&value)
			assert.True(t, queue.Peek())

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 5, value.Value)
			assert.True(t, queue.Dequeue(validatedValue))
			assert.EqualValues(t, 5, value.Value)

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 8, value.Value)
			assert.True(t, queue.Dequeue())

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 9, value.Value)
			assert.True(t, queue.Dequeue(validatedValue))
			assert.EqualValues(t, 9, value.Value)

			assert.False(t, queue.Peek())
			assert.False(t, queue.Peek(validatedValue))
			assert.False(t, queue.Dequeue(validatedValue))
		}).Run(t)
}
