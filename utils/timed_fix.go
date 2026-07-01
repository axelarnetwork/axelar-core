package utils

import (
	"os"
	"strings"
	"time"
)

// Activation times for consensus-breaking bug fixes per chain.
// Can be overridden via FIX_ACTIVATION_TIME environment variable (RFC3339 format).
const (
	MainnetFixActivationTime  = "2026-07-07T08:00:00Z"
	TestnetFixActivationTime  = "2026-07-03T08:00:00Z"
	StagenetFixActivationTime = "2026-07-02T12:00:00Z"
	DevnetFixActivationTime   = "2026-07-02T12:00:00Z"
)

func getFixActivationTime(chainID string) string {
	if envVal := os.Getenv("FIX_ACTIVATION_TIME"); envVal != "" {
		return envVal
	}

	if strings.Contains(chainID, "devnet") {
		return DevnetFixActivationTime
	}
	if strings.HasPrefix(chainID, "axelar-stagenet") {
		return StagenetFixActivationTime
	}
	if strings.HasPrefix(chainID, "axelar-testnet") {
		return TestnetFixActivationTime
	}
	// Default to mainnet activation time for axelar-dojo-1 and any unknown chain
	return MainnetFixActivationTime
}

// IsFixActive reports whether a consensus-breaking bug fix is active for the
// given chain at blockTime. Gating on block time (rather than deploying the new
// code unconditionally) lets every validator switch behavior at the same height,
// preserving consensus across the upgrade.
func IsFixActive(chainID string, blockTime time.Time) bool {
	activationTime, err := time.Parse(time.RFC3339, getFixActivationTime(chainID))
	if err != nil {
		return true // if parsing fails, activate the fix
	}
	return !blockTime.Before(activationTime)
}
