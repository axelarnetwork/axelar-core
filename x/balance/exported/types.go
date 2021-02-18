package exported

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Chain represents the properties of a registered blockchain
type Chain struct {
	Name                  string
	NativeAsset           string
	SupportsForeignAssets bool
}

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (c Chain) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("missing chain name")
	}
	if c.NativeAsset == "" {
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
