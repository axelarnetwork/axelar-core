package types

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState is the constructor of GenesisState
func NewGenesisState(
	params Params,
	nonce uint64,
	chains []exported.Chain,
	chainStates []ChainState,
	linkedAddresses []LinkedAddresses,
	transfers []exported.CrossChainTransfer,
	fee exported.TransferFee,
) *GenesisState {
	return &GenesisState{
		Params:          params,
		Nonce:           nonce,
		Chains:          chains,
		ChainStates:     chainStates,
		LinkedAddresses: linkedAddresses,
		Transfers:       transfers,
		Fee:             fee,
	}
}

// DefaultGenesisState creates the default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		0,
		[]exported.Chain{evm.Ethereum, axelarnet.Axelarnet},
		[]ChainState{{
			Chain:  axelarnet.Axelarnet,
			Assets: []exported.Asset{exported.NewAsset(axelarnet.NativeAsset, sdk.NewInt(100000), true)},
		}},
		[]LinkedAddresses{},
		[]exported.CrossChainTransfer{},
		exported.TransferFee{},
	)
}

// Validate checks if the genesis state is valid
func (m GenesisState) Validate() error {
	if err := m.Params.Validate(); err != nil {
		return getValidateError(err)
	}

	for _, chain := range m.Chains {
		if err := chain.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, chainState := range m.ChainStates {
		if err := chainState.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, linkedAddresses := range m.LinkedAddresses {
		if err := linkedAddresses.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	for _, transfer := range m.Transfers {
		if err := transfer.Validate(); err != nil {
			return getValidateError(err)
		}
	}

	if err := m.Fee.Coins.Validate(); err != nil {
		return getValidateError(err)
	}

	return nil
}

// GetGenesisStateFromAppState returns x/nexus GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}
