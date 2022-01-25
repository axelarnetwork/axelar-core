package types

import (
	fmt "fmt"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

func (m ChainState) indexOfAsset(asset string) int {
	for i := range m.Assets {
		if m.Assets[i].Denom == asset {
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

	for _, asset := range m.NativeAssets {
		if err := sdk.ValidateDenom(asset); err != nil {
			return sdkerrors.Wrap(err, "invalid native asset")
		}
	}

	if len(m.Assets) == 0 {
		return fmt.Errorf("no assets found")
	}

	seenDenoms := make(map[string]bool)

	for _, asset := range m.Assets {
		if !asset.IsNativeAsset && !m.Chain.SupportsForeignAssets {
			return fmt.Errorf("chain does not support foreign assets")
		}

		if err := asset.Validate(); err != nil {
			return sdkerrors.Wrap(err, "invalid asset")
		}

		if seenDenoms[asset.Denom] {
			return fmt.Errorf("duplicate asset found")
		}

		seenDenoms[asset.Denom] = true
	}

	return nil
}

// HasAsset returns true if the chain state has the given asset registered; false otherwise
func (m ChainState) HasAsset(asset string) bool {
	return m.indexOfAsset(asset) != -1
}

// AddAsset registers the given asset in chain state
func (m *ChainState) AddAsset(asset exported.Asset) error {
	if m.HasAsset(asset.Denom) {
		return fmt.Errorf("asset %s is already registered for chain %s", asset.Denom, m.Chain.Name)
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

// AssetMinAmount returns the minimum transfer amount for the chain
func (m ChainState) AssetMinAmount(asset string) sdk.Int {
	i := m.indexOfAsset(asset)
	if i == -1 {
		return sdk.ZeroInt()
	}

	return m.Assets[i].MinAmount
}

// AddNativeAsset registers the native asset for the chain
func (m *ChainState) AddNativeAsset(asset string) error {
	if m.HasNativeAsset(asset) {
		return fmt.Errorf("native asset %s is already registered for chain %s", asset, m.Chain.Name)
	}
	m.NativeAssets = append(m.NativeAssets, asset)

	return nil
}

// HasNativeAsset returns true if the chain has the given native asset ; false otherwise
func (m ChainState) HasNativeAsset(asset string) bool {
	return utils.IndexOf(m.NativeAssets, asset) != -1
}
