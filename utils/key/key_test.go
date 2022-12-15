package key_test

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/utils/slices"
)

func TestFromBz(t *testing.T) {
	expectedBz := []byte("testkey")
	k := key.FromBz(expectedBz)

	assert.Equal(t, expectedBz, k.Bytes())
}

func TestFromInt(t *testing.T) {
	k1 := key.FromUInt[uint64](0)
	k2 := key.FromUInt[uint64](17)
	k3 := key.FromUInt[uint64](17)
	k4 := key.FromUInt[uint64](math.MaxUint64)

	assert.True(t, bytes.Compare(k1.Bytes(), k2.Bytes()) < 0)
	assert.True(t, bytes.Compare(k2.Bytes(), k3.Bytes()) == 0)
	assert.True(t, bytes.Compare(k3.Bytes(), k4.Bytes()) < 0)

	assert.Equal(t, len(k1.Bytes()), len(k2.Bytes()))
	assert.Equal(t, len(k2.Bytes()), len(k3.Bytes()))
	assert.Equal(t, len(k3.Bytes()), len(k4.Bytes()))
}

func TestFromStr(t *testing.T) {
	expected := "testkey"
	k := key.FromStr(expected)

	assert.Equal(t, expected, string(k.Bytes()))
}

func TestAppend(t *testing.T) {
	k1 := key.FromStr("prefix")
	k2 := key.FromStr("nucleus")
	k3 := key.FromBz([]byte("suffix"))

	assert.Equal(t, []byte("prefix_nucleus_suffix"), k1.Append(k2).Append(k3).Bytes())
}

func TestRegisterPrefix(t *testing.T) {
	var keys []key.Key

	assert.NotPanics(t, func() { keys = append(keys, key.RegisterStaticKey("test_module", 1)) })
	assert.NotPanics(t, func() { keys = append(keys, key.RegisterStaticKey("test_module", 2)) })
	assert.NotPanics(t, func() { keys = append(keys, key.RegisterStaticKey("test_module", 5)) })

	assert.EqualValues(t, slices.Map(keys, key.Key.String), slices.Distinct(slices.Map(keys, key.Key.String)))

	var otherKey key.Key
	assert.NotPanics(t, func() { otherKey = key.RegisterStaticKey("other_module", 2) })
	assert.Equal(t, keys[1], otherKey)

	assert.Panics(t, func() { key.RegisterStaticKey("test_module", 2) })

}
