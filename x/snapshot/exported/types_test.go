package exported_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
)

func TestGetValidatorIllegibilities(t *testing.T) {
	expected := []exported.ValidatorIllegibility{exported.Tombstoned, exported.Jailed, exported.MissedTooManyBlocks, exported.NoProxyRegistered, exported.TssSuspended, exported.ProxyInsuficientFunds}
	actual := exported.GetValidatorIllegibilities()

	assert.Equal(t, expected, actual)
}

func TestFilterIllegibilityForNewKey(t *testing.T) {
	for _, illegibility := range exported.GetValidatorIllegibilities() {
		actual := illegibility.FilterIllegibilityForNewKey()

		assert.NotEqual(t, exported.None, actual)
	}
}

func TestFilterIllegibilityForTssSigning(t *testing.T) {
	for _, illegibility := range exported.GetValidatorIllegibilities() {
		actual := illegibility.FilterIllegibilityForTssSigning()

		assert.NotEqual(t, exported.None, actual)
	}
}

func TestFilterIllegibilityForMultisigSigning(t *testing.T) {
	for _, illegibility := range exported.GetValidatorIllegibilities() {
		actual := illegibility.FilterIllegibilityForMultisigSigning()

		switch illegibility {
		case exported.MissedTooManyBlocks:
			assert.Equal(t, exported.None, actual)
		case exported.ProxyInsuficientFunds:
			assert.Equal(t, exported.None, actual)
		default:
			assert.NotEqual(t, exported.None, actual)
		}
	}
}
