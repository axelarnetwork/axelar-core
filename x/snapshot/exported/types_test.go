package exported_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

func TestGetValidatorIllegibilities(t *testing.T) {
	expected := []exported.ValidatorIllegibility{exported.Tombstoned, exported.Jailed, exported.MissedTooManyBlocks, exported.NoProxyRegistered, exported.TssSuspended}
	actual := exported.GetValidatorIllegibilities()

	assert.Equal(t, expected, actual)
}
