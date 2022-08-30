package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

const maxBitmapSize = 1 << 15 // 32,768

// NewMaintainerState is the constructor for MaintainerState
func NewMaintainerState(address sdk.ValAddress) MaintainerState {
	return MaintainerState{
		Address:        address,
		MissingVotes:   utils.NewBitmap(maxBitmapSize),
		IncorrectVotes: utils.NewBitmap(maxBitmapSize),
	}
}

// ValidateBasic returns error if the given MaintainerState is invalid, nil otherwise
func (m MaintainerState) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Address); err != nil {
		return err
	}

	return nil
}

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

// IndexOfMaintainer returns the index of the maintainer in the given chain state
func (m ChainState) IndexOfMaintainer(address sdk.ValAddress) int {
	for i, ms := range m.MaintainerStates {
		if ms.Address.Equals(address) {
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

	for _, ms := range m.MaintainerStates {
		if err := ms.ValidateBasic(); err != nil {
			return err
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
func (m ChainState) HasMaintainer(address sdk.ValAddress) bool {
	return m.IndexOfMaintainer(address) != -1
}

// AddMaintainer registers the given maintainer for the chain
func (m *ChainState) AddMaintainer(address sdk.ValAddress) error {
	if m.HasMaintainer(address) {
		return fmt.Errorf("maintainer %s is already registered for chain %s", address.String(), m.Chain.Name)
	}

	m.MaintainerStates = append(m.MaintainerStates, NewMaintainerState(address))

	return nil
}

// RemoveMaintainer deregisters the given maintainer for the chain
func (m *ChainState) RemoveMaintainer(address sdk.ValAddress) error {
	i := m.IndexOfMaintainer(address)
	if i == -1 {
		return fmt.Errorf("maintainer %s is not registered for chain %s", address.String(), m.Chain.Name)
	}

	m.MaintainerStates[i] = m.MaintainerStates[len(m.MaintainerStates)-1]
	m.MaintainerStates = m.MaintainerStates[:len(m.MaintainerStates)-1]

	return nil
}

// MarkMissingVote marks the given chain maintainer for missing vote of a poll
func (m *ChainState) MarkMissingVote(address sdk.ValAddress, missingVote bool) {
	i := m.IndexOfMaintainer(address)
	if i == -1 {
		return
	}

	m.MaintainerStates[i].MissingVotes.Add(missingVote)
}

// MarkIncorrectVote marks the given chain maintainer for voting incorrectly of a poll
func (m *ChainState) MarkIncorrectVote(address sdk.ValAddress, incorrectVote bool) {
	i := m.IndexOfMaintainer(address)
	if i == -1 {
		return
	}

	m.MaintainerStates[i].IncorrectVotes.Add(incorrectVote)
}

// ChainName returns the chain name for the given state
func (m ChainState) ChainName() exported.ChainName {
	return m.Chain.Name
}
