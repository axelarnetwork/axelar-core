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
func NewMaintainerState(chain exported.ChainName, address sdk.ValAddress) *MaintainerState {
	return &MaintainerState{
		Chain:          chain,
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

// ChainName returns the chain name for the given state
func (m ChainState) ChainName() exported.ChainName {
	return m.Chain.Name
}

var _ exported.MaintainerState = &MaintainerState{}

// MarkMissingVote marks the given maintainer for missing vote of a poll
func (m *MaintainerState) MarkMissingVote(missingVote bool) {
	m.MissingVotes.Add(missingVote)
}

// MarkIncorrectVote marks the given maintainer for voting incorrectly of a poll
func (m *MaintainerState) MarkIncorrectVote(incorrectVote bool) {
	m.IncorrectVotes.Add(incorrectVote)
}

// CountMissingVotes returns the number of missing votes within the given window
func (m MaintainerState) CountMissingVotes(window int) uint64 {
	return m.MissingVotes.CountTrue(window)
}

// CountIncorrectVotes returns the number of incorrect votes within the given window
func (m MaintainerState) CountIncorrectVotes(window int) uint64 {
	return m.IncorrectVotes.CountTrue(window)
}

// GetAddress returns the address of the maintainer
func (m MaintainerState) GetAddress() sdk.ValAddress { return m.Address }
