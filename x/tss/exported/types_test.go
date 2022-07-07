package exported_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssTestUtils "github.com/axelarnetwork/axelar-core/x/tss/exported/testutils"
)

func TestKeyID_Validate(t *testing.T) {
	repeats := 20
	t.Run("GIVEN a valid key ID WHEN validating THEN return no error", testutils.Func(func(t *testing.T) {
		keyID := exported.KeyID(rand.NormalizedStrBetween(exported.KeyIDLengthMin, exported.KeyIDLengthMax+1))
		assert.NoError(t, keyID.Validate())
	}).Repeat(repeats))

	t.Run("GIVEN a short key ID WHEN validating THEN return error", testutils.Func(func(t *testing.T) {
		keyID := exported.KeyID(rand.NormalizedStrBetween(1, exported.KeyIDLengthMin))
		assert.Error(t, keyID.Validate())
	}).Repeat(repeats))

	t.Run("GIVEN a long key ID WHEN validating THEN return error", testutils.Func(func(t *testing.T) {
		keyID := exported.KeyID(rand.NormalizedStrBetween(exported.KeyIDLengthMax+1, 2*exported.KeyIDLengthMax+1))
		assert.Error(t, keyID.Validate())
	}).Repeat(repeats))

	t.Run("GIVEN a key ID with separator WHEN validating THEN return error", testutils.Func(func(t *testing.T) {
		keyID := exported.KeyID(strings.Repeat("_", exported.KeyIDLengthMin))
		assert.Error(t, keyID.Validate())
	}).Repeat(repeats))
}

func TestKeyIDsToStrings(t *testing.T) {
	repeats := 5
	t.Run("GIVEN a slice of key IDs WHEN converting THEN return equivalent slice of strings", testutils.Func(func(t *testing.T) {
		keyIDs := make([]exported.KeyID, 0, rand.I64Between(1, 20))
		for i := 0; i < cap(keyIDs); i++ {
			keyIDs = append(keyIDs, tssTestUtils.RandKeyID())
		}

		strs := exported.KeyIDsToStrings(keyIDs)

		for i := range keyIDs {
			assert.Equal(t, string(keyIDs[i]), strs[i])
		}
	}).Repeat(repeats))

	t.Run("GIVEN an empty slice of key IDs WHEN converting THEN return an empty slice of strings", func(t *testing.T) {
		keyIDs := make([]exported.KeyID, 0)
		strs := exported.KeyIDsToStrings(keyIDs)
		assert.Equal(t, []string{}, strs)
	})

	t.Run("GIVEN a nil slice of key IDs WHEN converting THEN return a nil slice of strings", func(t *testing.T) {
		var keyIDs []exported.KeyID
		strs := exported.KeyIDsToStrings(keyIDs)
		assert.Nil(t, strs)
	})
}

func TestComputeAbsCorruptionThreshold(t *testing.T) {
	assert.Equal(t, int64(7), exported.ComputeAbsCorruptionThreshold(utils.NewThreshold(2, 3), sdk.NewInt(12)))
	assert.Equal(t, int64(3), exported.ComputeAbsCorruptionThreshold(utils.NewThreshold(11, 20), sdk.NewInt(7)))
}
