package exported

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

// CrossChainAddress represents a generalized address on any registered chain
type CrossChainAddress struct {
	Chain   Chain
	Address string
}

// Validate performs a stateless check to ensure the CrossChainAddress object has been initialized correctly
func (a CrossChainAddress) Validate() error {
	if err := a.Chain.Validate(); err != nil {
		return err
	}
	if a.Address == "" {
		return fmt.Errorf("missing address")
	}
	return nil
}

func (a CrossChainAddress) String() string {
	return fmt.Sprintf("chain: %s, address: %s", a.Chain.Name, a.Address)
}

// CrossChainTransfer represents a generalized transfer of some asset to a registered blockchain
type CrossChainTransfer struct {
	Recipient CrossChainAddress
	Asset     sdk.Coin
	ID        uint64
}
