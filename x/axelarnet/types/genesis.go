package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
)

// NewGenesisState returns a new GenesisState instance
func NewGenesisState(p Params, feeCollector sdk.AccAddress, chains []CosmosChain, transferQueue utils.QueueState) *GenesisState {
	SortChains(chains)
	return &GenesisState{
		Params:           p,
		CollectorAddress: feeCollector,
		Chains:           chains,
		TransferQueue:    transferQueue,
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
		TransferQueue: utils.QueueState{},
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

	if err := m.TransferQueue.ValidateBasic(); err != nil {
		return getValidateError(sdkerrors.Wrapf(err, "invalid transfer queue state"))
	}

	return nil
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}
