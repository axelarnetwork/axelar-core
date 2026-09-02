package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func parseActivation(t *testing.T, activationTime string) time.Time {
	parsed, err := time.Parse(time.RFC3339, activationTime)
	assert.NoError(t, err)

	return parsed
}

// isActive panics on a malformed constant, so make a bad edit fail here instead
func TestActivationTimesParse(t *testing.T) {
	for name, a := range map[string]activations{"v1.5.3": v153} {
		t.Run(name, func(t *testing.T) {
			for _, activationTime := range []string{a.mainnet, a.testnet, a.stagenet, a.devnet} {
				parseActivation(t, activationTime)
			}
		})
	}
}

func TestIsV153Active(t *testing.T) {
	stagenet := parseActivation(t, v153.stagenet)
	devnet := parseActivation(t, v153.devnet)

	t.Run("mainnet and testnet are active from their first block", func(t *testing.T) {
		// mainnet runs the v1.5 upgrade on this binary, testnet ran it on v1.5.3
		for _, chainID := range []string{"axelar-dojo-1", "axelar-testnet-lisbon-3", "", "some-other-chain-1"} {
			assert.True(t, IsV153Active(chainID, time.Time{}))
		}
	})

	t.Run("inactive strictly before, active at and after the activation time", func(t *testing.T) {
		assert.False(t, IsV153Active("axelar-stagenet-724", stagenet.Add(-time.Nanosecond)))
		assert.True(t, IsV153Active("axelar-stagenet-724", stagenet))
		assert.False(t, IsV153Active("devnet-amplifier", devnet.Add(-time.Nanosecond)))
		assert.True(t, IsV153Active("devnet-amplifier", devnet))
	})

	t.Run("re-genesised networks resolve to their own entry", func(t *testing.T) {
		assert.False(t, IsV153Active("axelar-stagenet-825", stagenet.Add(-time.Nanosecond)))
		assert.False(t, IsV153Active("devnet-verifiers", devnet.Add(-time.Nanosecond)))
		assert.True(t, IsV153Active("axelar-testnet-casablanca-1", time.Time{}))
	})
}
