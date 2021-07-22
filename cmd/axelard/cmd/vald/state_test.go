package vald_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/mock"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestStateStore_GetState(t *testing.T) {
	repeats := 20
	rw := &mock.ReadWriterMock{}
	store := vald.NewStateStore(rw)

	t.Run("return positive block height", testutils.Func(func(t *testing.T) {
		expected := rand.PosI64()
		rw.ReadAllFunc = func() ([]byte, error) { return json.Marshal(expected) }
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("return block 0", testutils.Func(func(t *testing.T) {
		expected := int64(0)
		rw.ReadAllFunc = func() ([]byte, error) { return json.Marshal(expected) }
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("negative value", testutils.Func(func(t *testing.T) {
		expected := -1 * rand.PosI64()
		rw.ReadAllFunc = func() ([]byte, error) { return json.Marshal(expected) }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("reader error", testutils.Func(func(t *testing.T) {
		rw.ReadAllFunc = func() ([]byte, error) { return nil, fmt.Errorf("some error") }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("wrong data format", testutils.Func(func(t *testing.T) {
		rw.ReadAllFunc = func() ([]byte, error) { return rand.BytesBetween(1, 100), nil }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))

	t.Run("empty reader", testutils.Func(func(t *testing.T) {
		rw.ReadAllFunc = func() ([]byte, error) { return []byte{}, nil }
		_, err := store.GetState()
		assert.Error(t, err)
	}).Repeat(repeats))
}

func TestStateStore_SetState(t *testing.T) {
	repeats := 20
	rw := &mock.ReadWriterMock{}
	store := vald.NewStateStore(rw)

	t.Run("persist positive block height", testutils.Func(func(t *testing.T) {
		var storage []byte
		rw.ReadAllFunc = func() ([]byte, error) { return storage, nil }
		rw.WriteAllFunc = func(bz []byte) error { storage = bz; return nil }
		expected := rand.PosI64()
		assert.NoError(t, store.SetState(expected))
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("persist block 0", testutils.Func(func(t *testing.T) {
		var storage []byte
		rw.ReadAllFunc = func() ([]byte, error) { return storage, nil }
		rw.WriteAllFunc = func(bz []byte) error { storage = bz; return nil }
		expected := int64(0)
		assert.NoError(t, store.SetState(expected))
		actual, err := store.GetState()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}).Repeat(repeats))

	t.Run("negative value", testutils.Func(func(t *testing.T) {
		rw.WriteAllFunc = func(bz []byte) error { return nil }
		assert.Error(t, store.SetState(-1*rand.PosI64()))
	}).Repeat(repeats))

	t.Run("write fails", testutils.Func(func(t *testing.T) {
		rw.WriteAllFunc = func(bz []byte) error { return fmt.Errorf("some error") }
		assert.Error(t, store.SetState(rand.PosI64()))
	}).Repeat(repeats))
}
