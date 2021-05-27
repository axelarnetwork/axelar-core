package exported

import (
	"fmt"
	"strings"
)

const (

	// Known platforms (TODO: ethereum will eventually be renamed to EVM)
	ethPlatform = "ethereum"
	btcPlatform = "bitcoin"

	// Hardcoded chains
	bitcoin = "bitcoin"
)

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (m Chain) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("missing chain name")
	}
	if m.NativeAsset == "" {
		return fmt.Errorf("missing native asset name")
	}

	// check hardcoded chains
	switch strings.ToLower(m.Name) {

	// The more hardcoded chains we add in the future, the more
	// cases will be needed here
	case bitcoin:
		return fmt.Errorf("bitcoin is a hardcoded chain")
	}

	switch strings.ToLower(m.Platform) {

	// The more platforms we add in the future, the more
	// cases will be needed here
	case ethPlatform:
	default:
		return fmt.Errorf("unknown or reserved platform")
	}
	return nil
}
