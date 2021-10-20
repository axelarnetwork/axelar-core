package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m ChainState) indexOfMaintainer(maintainer sdk.ValAddress) int {
	for i, mt := range m.Maintainers {
		if mt.Equals(maintainer) {
			return i
		}
	}

	return -1
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
