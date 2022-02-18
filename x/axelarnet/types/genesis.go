package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState returns a new GenesisState instance
func NewGenesisState(p Params, feeCollector sdk.AccAddress, chains []CosmosChain, transfers []IBCTransfer) *GenesisState {
	SortChains(chains)
	SortTransfers(transfers)
	return &GenesisState{
		Params:           p,
		CollectorAddress: feeCollector,
		Chains:           chains,
		PendingTransfers: transfers,
	}
}

// DefaultGenesisState represents the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:           DefaultParams(),
		CollectorAddress: nil,
		Chains: []CosmosChain{{
			Name:       exported.Axelarnet.Name,
			AddrPrefix: "axelar",
		}},
		PendingTransfers: []IBCTransfer{},
	}
}

// Validate checks if the genesis state is valid
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
	}

	if len(m.CollectorAddress) > 0 {
		if err := sdk.VerifyAddressFormat(m.CollectorAddress); err != nil {
			return getValidateError(err)
		}
	}

	for i, chain := range m.Chains {
		if err := chain.Validate(); err != nil {
			return getValidateError(sdkerrors.Wrap(err, fmt.Sprintf("faulty chain entry %d", i)))
		}
	}

	for i, transfer := range m.PendingTransfers {
		if err := transfer.Validate(); err != nil {
			return getValidateError(sdkerrors.Wrap(err, fmt.Sprintf("faulty transfer entry %d", i)))
		}
	}

	return nil
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}
