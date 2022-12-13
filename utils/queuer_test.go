package utils_test

import (
	"errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"golang.org/x/crypto/sha3"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestQueue(t *testing.T) {
	var (
		queue utils.Queue[utils.ValidatedProtoMarshaler]
	)

	Given("a queue without prioritization", func() {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		cdc := app.MakeEncodingConfig().Codec
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("storeKey")), cdc)
		queue = utils.NewQueue(store, key.FromStrHashed("queue1"), func(_ utils.ValidatedProtoMarshaler) key.Key { return key.FromRaw([]byte{}) })
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
			iter := queue.Iter()
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
			iter := queue.Iter()
			defer func() { _ = iter.Close() }()

			assert.False(t, iter.Valid())
		}).Run(t)

	Given("a queue with prioritization", func() {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		cdc := app.MakeEncodingConfig().Codec
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("storeKey")), cdc)

		prio := uint(10)
		queue = utils.NewQueue(store, key.FromStrHashed("queue1"), func(v utils.ValidatedProtoMarshaler) key.Key {
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
			assert.True(t, queue.Any())

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 5, value.Value)
			assert.True(t, queue.Dequeue(validatedValue))
			assert.EqualValues(t, 5, value.Value)

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 8, value.Value)
			assert.True(t, queue.Discard())

			assert.True(t, queue.Peek(validatedValue))
			assert.EqualValues(t, 9, value.Value)
			assert.True(t, queue.Dequeue(validatedValue))
			assert.EqualValues(t, 9, value.Value)

			assert.False(t, queue.Any())
			assert.False(t, queue.Peek(validatedValue))
			assert.False(t, queue.Dequeue(validatedValue))
		}).Run(t)
}

func TestIndexedQueue(t *testing.T) {
	var (
		queue utils.IndexedQueue[utils.ValidatedProtoMarshaler, []int32]
	)

	Given("a queue without prioritization", func() {
		ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		cdc := app.MakeEncodingConfig().Codec
		store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("storeKey")), cdc)
		q := utils.NewQueue(store, key.FromStrHashed("queue1"), func(_ utils.ValidatedProtoMarshaler) key.Key { return key.FromRaw([]byte{}) })
		queue = utils.NewIndexedQueue(q, func(value utils.ValidatedProtoMarshaler) []int32 {
			return hashHelper(value)
		})
	}).
		When("enqueuing items", func() {
			assert.NoError(t, queue.Enqueue(itemize(9)))
			assert.NoError(t, queue.Enqueue(itemize(8)))
			assert.Error(t, queue.Enqueue(utils.WithValidation(&gogoprototypes.UInt64Value{Value: 6}, func() error {
				return errors.New("wrong input")
			})))
			assert.NoError(t, queue.Enqueue(itemize(5)))
		}).
		Branch(
			Then("we can check existence of an item", func(t *testing.T) {
				assert.True(t, queue.Has(hashHelper(itemize(9))))
				assert.False(t, queue.Has(hashHelper(itemize(6))))
			}).
				Then("we can lookup an item", func(t *testing.T) {
					var value gogoprototypes.UInt64Value
					validatedValue := utils.NoValidation(&value)
					assert.True(t, queue.Get(hashHelper(itemize(8)), validatedValue))
					assert.False(t, queue.Get(hashHelper(itemize(6)), validatedValue))
				}),
			When("dequeuing an item", func() {
				var value gogoprototypes.UInt64Value
				validatedValue := utils.NoValidation(&value)
				assert.True(t, queue.Has(hashHelper(itemize(9))))
				assert.True(t, queue.Dequeue(validatedValue))
			}).
				Then("the index of that item is erased", func(t *testing.T) {
					assert.False(t, queue.Has(hashHelper(itemize(9))))
				})).Run(t)
}

func itemize(value uint64) utils.ValidatedProtoMarshaler {
	return utils.NoValidation(&gogoprototypes.UInt64Value{Value: value})
}

func hashHelper(value utils.ValidatedProtoMarshaler) []int32 {
	hash := sha3.Sum256(funcs.Must(value.Marshal()))
	return slices.TryCast[byte, int32](hash[:])
}
