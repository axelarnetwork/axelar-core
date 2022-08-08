package utils

import (
	"bytes"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	abci "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils/key"
)

func TestKey(t *testing.T) {
	repeats := 20
	t.Run("all lower case", testutils.Func(func(t *testing.T) {
		keyStr := rand.StrBetween(1, 30)
		lck := LowerCaseKey(keyStr)

		assert.Equal(t, []byte(strings.ToLower(keyStr)), lck.AsKey())
	}).Repeat(repeats))

	t.Run("lower case key equals key with ToLower transformation", testutils.Func(func(t *testing.T) {
		keyStr := rand.StrBetween(1, 30)
		lck1 := KeyFromStr(keyStr, strings.ToLower)
		lck2 := LowerCaseKey(keyStr)

		assert.True(t, bytes.Equal(lck1.AsKey(), lck2.AsKey()))
	}).Repeat(repeats))

	t.Run("different keys", testutils.Func(func(t *testing.T) {
		key1 := KeyFromStr(rand.StrBetween(1, 30))
		key2 := KeyFromStr(rand.StrBetween(1, 30))

		assert.False(t, bytes.Equal(key1.AsKey(), key2.AsKey()))
	}).Repeat(repeats))

	t.Run("prepends creates same key as append", testutils.Func(func(t *testing.T) {
		key1 := KeyFromStr(rand.StrBetween(1, 30))
		key2 := KeyFromStr(rand.StrBetween(1, 30))
		key3 := KeyFromStr(rand.StrBetween(1, 30))
		key4 := KeyFromStr(rand.StrBetween(1, 30))
		key5 := KeyFromStr(rand.StrBetween(1, 30))
		compKey1 := key1.Append(key2).Append(key3).Append(key4).Append(key5)
		compKey2 := key5.Prepend(key4).Prepend(key3).Prepend(key2).Prepend(key1)

		assert.True(t, bytes.Equal(compKey1.AsKey(), compKey2.AsKey()))
	}).Repeat(repeats))

	t.Run("key from integer", func(t *testing.T) {
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0x10}, KeyFromInt(16).AsKey())
	})
}

func TestKVStore_Get(t *testing.T) {
	encConf := params.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("test")), encConf.Codec)

	filledState := QueueState{
		Items: map[string]QueueState_Item{"state": {Key: []byte("stateKey1"), Value: []byte("stateValue1")}},
	}
	emptyState := QueueState{}

	store.Set(KeyFromStr("key"), &emptyState)

	assert.True(t, store.Get(KeyFromStr("key"), &filledState))
	assert.Equal(t, emptyState, filledState)
}

func TestKVStore_GetNew(t *testing.T) {
	encConf := params.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("test")), encConf.Codec)

	filledState := QueueState{
		Items: map[string]QueueState_Item{"state": {Key: []byte("stateKey1"), Value: []byte("stateValue1")}},
	}
	emptyState := QueueState{}

	store.SetNew(key.FromStr("key"), &emptyState)

	assert.True(t, store.GetNew(key.FromStr("key"), &filledState))
	assert.Equal(t, emptyState, filledState)
}

func TestKVStore_Iterator(t *testing.T) {
	encConf := params.MakeEncodingConfig()
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	store := NewNormalizedStore(ctx.KVStore(sdk.NewKVStoreKey("test")), encConf.Codec)

	filledState := QueueState{
		Items: map[string]QueueState_Item{"state": {Key: []byte("stateKey1"), Value: []byte("stateValue1")}},
	}
	emptyState := QueueState{}

	storeKey := KeyFromStr("prefix_key")
	store.Set(storeKey, &emptyState)

	iter := store.Iterator(KeyFromStr("prefix"))

	assert.True(t, iter.Valid())
	iter.UnmarshalValue(&filledState)

	assert.Equal(t, emptyState, filledState)
}

func Test_Reset(t *testing.T) {
	var state QueueState
	assert.NotPanics(t, func() {
		(&state).Reset()
	})

	assert.NotPanics(t, func() {
		state = QueueState{}
		(&state).Reset()
	})
}
