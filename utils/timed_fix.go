package utils

import (
	"strings"
	"time"

	"github.com/axelarnetwork/utils/funcs"
)

const alwaysActive = "0001-01-01T00:00:00Z"

type activations struct {
	mainnet, testnet, stagenet, devnet string
}

// v153 gates the state breaking changes released in v1.5.3 for chains that
// already ran the v1.5 upgrade on v1.5.2, currently the gas charged for the
// SubmitPubKey proof of ownership verification. Mainnet and testnet run the v1.5
// upgrade on v1.5.3, which already charges the gas, so it is in effect there from
// their first v1.5 block; stagenet and devnet ran v1.5 on v1.5.2, which does not.
var v153 = activations{
	mainnet:  alwaysActive,
	testnet:  alwaysActive,
	stagenet: "2026-10-20T08:00:00Z",
	devnet:   "2026-10-20T08:00:00Z",
}

func (a activations) isActive(chainID string, blockTime time.Time) bool {
	activationTime := a.mainnet
	switch {
	case strings.Contains(chainID, "devnet"):
		activationTime = a.devnet
	case strings.HasPrefix(chainID, "axelar-stagenet"):
		activationTime = a.stagenet
	case strings.HasPrefix(chainID, "axelar-testnet"):
		activationTime = a.testnet
	}

	return !blockTime.Before(funcs.Must(time.Parse(time.RFC3339, activationTime)))
}

// IsV153Active reports whether the v1.5.3 changes are in effect for the given
// chain at blockTime.
func IsV153Active(chainID string, blockTime time.Time) bool {
	return v153.isActive(chainID, blockTime)
}
