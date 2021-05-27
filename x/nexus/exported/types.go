package exported

import (
	"fmt"
	"strings"
)

var (
	// Hardcoded chains in axelar-core
	hardcoded = []string{"bitcoin"}
)

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (m Chain) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("missing chain name")
	}
	if m.NativeAsset == "" {
		return fmt.Errorf("missing native asset name")
	}

	return nil
}

// IsHardCoded returns true if the chain is one that is hardcoded in axelar-core
func (m Chain) IsHardCoded() bool {
	for _, c := range hardcoded {
		if strings.ToLower(m.Name) == c {
			return true
		}
	}

	return false
}
