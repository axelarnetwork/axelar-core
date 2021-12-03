package types

import (
	fmt "fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewLinkedAddresses is the constructor of LinkedAddresses
func NewLinkedAddresses(depositAddress, recepientAddress exported.CrossChainAddress) LinkedAddresses {
	return LinkedAddresses{
		DepositAddress:   depositAddress,
		RecipientAddress: recepientAddress,
	}
}

// Validate validates the LinkedAddresses
func (m LinkedAddresses) Validate() error {
	if err := m.DepositAddress.Validate(); err != nil {
		return err
	}

	if err := m.RecipientAddress.Validate(); err != nil {
		return err
	}

	return nil
}

func (m ChainState) indexOfMaintainer(maintainer sdk.ValAddress) int {
	for i, mt := range m.Maintainers {
		if mt.Equals(maintainer) {
			return i
		}
	}

	return -1
}

// Validate validates the ChainState
func (m ChainState) Validate() error {
	if err := m.Chain.Validate(); err != nil {
		return err
	}

	for _, maintainer := range m.Maintainers {
		if err := sdk.VerifyAddressFormat(maintainer); err != nil {
			return err
		}
	}

	if err := m.Total.Validate(); err != nil {
		return err
	}

	for _, asset := range m.Assets {
		if err := sdk.ValidateDenom(asset); err != nil {
			return err
		}
	}

	return nil
}

// HasAsset returns true if the chain state has the given asset registered; false otherwise
func (m ChainState) HasAsset(asset string) bool {
	return utils.IndexOf(m.Assets, asset) != -1
}

// AddAsset registers the given asset in chain state
func (m *ChainState) AddAsset(asset string) error {
	if m.HasAsset(asset) {
		return fmt.Errorf("asset %s is already registered for chain %s", asset, m.Chain.Name)
	}

	m.Assets = append(m.Assets, asset)

	return nil
}

// HasMaintainer returns true if the given maintainer is registered for the chain; false otherwise
func (m ChainState) HasMaintainer(maintainer sdk.ValAddress) bool {
	return m.indexOfMaintainer(maintainer) != -1
}

// AddMaintainer registers the given maintainer for the chain
func (m *ChainState) AddMaintainer(maintainer sdk.ValAddress) error {
	if m.HasMaintainer(maintainer) {
		return fmt.Errorf("maintainer %s is already registered for chain %s", maintainer.String(), m.Chain.Name)
	}

	m.Maintainers = append(m.Maintainers, maintainer)

	return nil
}

// RemoveMaintainer deregisters the given maintainer for the chain
func (m *ChainState) RemoveMaintainer(maintainer sdk.ValAddress) error {
	i := m.indexOfMaintainer(maintainer)
	if i == -1 {
		return fmt.Errorf("maintainer %s is not registered for chain %s", maintainer.String(), m.Chain.Name)
	}

	m.Maintainers[i] = m.Maintainers[len(m.Maintainers)-1]
	m.Maintainers = m.Maintainers[:len(m.Maintainers)-1]

	return nil
}
