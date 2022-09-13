package utils_test

import (
	"bytes"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/utils/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestKey(t *testing.T) {
	repeats := 20
	t.Run("all lower case", testutils.Func(func(t *testing.T) {
		keyStr := rand.StrBetween(1, 30)
		lck := utils.LowerCaseKey(keyStr)

		assert.Equal(t, []byte(strings.ToLower(keyStr)), lck.AsKey())
	}).Repeat(repeats))

	t.Run("lower case key equals key with ToLower transformation", testutils.Func(func(t *testing.T) {
		keyStr := rand.StrBetween(1, 30)
		lck1 := utils.KeyFromStr(keyStr, strings.ToLower)
		lck2 := utils.LowerCaseKey(keyStr)

		assert.True(t, bytes.Equal(lck1.AsKey(), lck2.AsKey()))
	}).Repeat(repeats))

	t.Run("different keys", testutils.Func(func(t *testing.T) {
		key1 := utils.KeyFromStr(rand.StrBetween(1, 30))
		key2 := utils.KeyFromStr(rand.StrBetween(1, 30))

		assert.False(t, bytes.Equal(key1.AsKey(), key2.AsKey()))
	}).Repeat(repeats))

	t.Run("prepends creates same key as append", testutils.Func(func(t *testing.T) {
		key1 := utils.KeyFromStr(rand.StrBetween(1, 30))
		key2 := utils.KeyFromStr(rand.StrBetween(1, 30))
		key3 := utils.KeyFromStr(rand.StrBetween(1, 30))
		key4 := utils.KeyFromStr(rand.StrBetween(1, 30))
		key5 := utils.KeyFromStr(rand.StrBetween(1, 30))
		compKey1 := key1.Append(key2).Append(key3).Append(key4).Append(key5)
		compKey2 := key5.Prepend(key4).Prepend(key3).Prepend(key2).Prepend(key1)

		assert.True(t, bytes.Equal(compKey1.AsKey(), compKey2.AsKey()))
	}).Repeat(repeats))

	t.Run("key from integer", func(t *testing.T) {
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0x10}, utils.KeyFromInt(16).AsKey())
	})
}

func TestKVStore_Get(t *testing.T) {
	encConf := params.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("test")), encConf.Codec)

	filledState := utils.QueueState{
		Items: map[string]utils.QueueState_Item{"state": {Key: []byte("stateKey1"), Value: []byte("stateValue1")}},
	}
	emptyState := utils.QueueState{}

	store.Set(utils.KeyFromStr("key"), &emptyState)

	assert.True(t, store.Get(utils.KeyFromStr("key"), &filledState))
	assert.Equal(t, emptyState, filledState)
}

func TestKVStore_GetNew(t *testing.T) {
	encConf := params.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("test")), encConf.Codec)

	filledState := utils.QueueState{
		Items: map[string]utils.QueueState_Item{"state": {Key: []byte("stateKey1"), Value: []byte("stateValue1")}},
	}
	emptyState := utils.QueueState{}

	assert.NoError(t, store.SetNewValidated(key.FromStr("key"), utils.NoValidation(&emptyState)))

	assert.True(t, store.GetNew(key.FromStr("key"), &filledState))
	assert.Equal(t, emptyState, filledState)
}

func TestKVStore_SetNewValidated(t *testing.T) {
	var (
		store utils.KVStore
		value *mock.ValidatedProtoMarshalerMock
	)

	givenKVStore := Given("a kv store", func() {
		encConf := params.MakeEncodingConfig()
		ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
		store = utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("test")), encConf.Codec)
	})

	givenKVStore.
		When("the value is invalid", func() {
			value = &mock.ValidatedProtoMarshalerMock{
				ValidateBasicFunc: func() error { return errors.New("some error") },
			}
		}).
		Then("fail to store the value", func(t *testing.T) {
			assert.Error(t, store.SetNewValidated(key.FromStr("key"), value))
		}).Run(t)

	givenKVStore.
		When("the value is valid", func() {
			value = &mock.ValidatedProtoMarshalerMock{
				ValidateBasicFunc: func() error { return nil },
				MarshalFunc:       func() ([]byte, error) { return []byte("marshaled"), nil },
				SizeFunc:          func() int { return len("marshaled") },
			}
		}).
		Then("store the value", func(t *testing.T) {
			assert.NoError(t, store.SetNewValidated(key.FromStr("key"), value))
		}).
		Then("return value", func(t *testing.T) {
			assert.NotPanics(t, func() {
				newValue := &mock.ValidatedProtoMarshalerMock{
					ResetFunc: func() {},
					UnmarshalFunc: func(bz []byte) error {
						if !bytes.Equal([]byte("marshaled"), bz) {
							return errors.New("unmarshal error")
						}
						return nil
					},
				}

				assert.True(t, store.GetNew(key.FromStr("key"), newValue))
			})
		},
		).Run(t)
}

func TestKVStore_Iterator(t *testing.T) {
	encConf := params.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	store := utils.NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("test")), encConf.Codec)

	filledState := utils.QueueState{
		Items: map[string]utils.QueueState_Item{"state": {Key: []byte("stateKey1"), Value: []byte("stateValue1")}},
	}
	emptyState := utils.QueueState{}

	storeKey := utils.KeyFromStr("prefix_key")
	store.Set(storeKey, &emptyState)

	iter := store.Iterator(utils.KeyFromStr("prefix"))

	assert.True(t, iter.Valid())
	iter.UnmarshalValue(&filledState)

	assert.Equal(t, emptyState, filledState)
}

func Test_Reset(t *testing.T) {
	var state utils.QueueState
	assert.NotPanics(t, func() {
		(&state).Reset()
	})

	assert.NotPanics(t, func() {
		state = utils.QueueState{}
		(&state).Reset()
	})
}
