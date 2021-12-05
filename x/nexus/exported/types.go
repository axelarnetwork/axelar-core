package exported

import (
	"fmt"
)

// AddressValidator defines a function that implements address verification upon a request to link addresses
type AddressValidator func(ctx sdk.Context, address CrossChainAddress) error

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (m Chain) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("missing chain name")
	}

	if m.NativeAsset == "" {
		return fmt.Errorf("missing native asset name")
	}

	if err := m.KeyType.Validate(); err != nil {
		return err
	}

	if m.Module == "" {
		return fmt.Errorf("missing module name")
	}

	return nil
}
