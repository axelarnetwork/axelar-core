package exported

import (
	"fmt"
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
