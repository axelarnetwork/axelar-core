package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
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

		assert.True(t, lck1.Equals(lck2))
	}).Repeat(repeats))

	t.Run("different keys", testutils.Func(func(t *testing.T) {
		key1 := KeyFromStr(rand.StrBetween(1, 30))
		key2 := KeyFromStr(rand.StrBetween(1, 30))

		assert.False(t, key1.Equals(key2))
	}).Repeat(repeats))

	t.Run("prepends creates same key as append", testutils.Func(func(t *testing.T) {
		key1 := KeyFromStr(rand.StrBetween(1, 30))
		key2 := KeyFromStr(rand.StrBetween(1, 30))
		key3 := KeyFromStr(rand.StrBetween(1, 30))
		key4 := KeyFromStr(rand.StrBetween(1, 30))
		key5 := KeyFromStr(rand.StrBetween(1, 30))
		compKey1 := key1.Append(key2).Append(key3).Append(key4).Append(key5)
		compKey2 := key5.Prepend(key4).Prepend(key3).Prepend(key2).Prepend(key1)

		assert.True(t, compKey1.Equals(compKey2))
	}).Repeat(repeats))

}
