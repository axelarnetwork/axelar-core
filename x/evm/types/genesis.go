package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewGenesisState returns a new genesis state
func NewGenesisState(chains []GenesisState_Chain) GenesisState {
	sort.Slice(chains, less(chains))

	return GenesisState{Chains: chains}
}

func less(chains []GenesisState_Chain) func(i int, j int) bool {
	return func(i, j int) bool {
		return chains[i].Params.Chain < chains[j].Params.Chain
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{Chains: DefaultChains()}
}

// DefaultChains returns the default chains for a genesis state
func DefaultChains() []GenesisState_Chain {
	p := DefaultParams()

	var chains []GenesisState_Chain
	for _, params := range p {
		chain := GenesisState_Chain{
			Params:            params,
			BurnerInfos:       nil,
			CommandQueue:      nil,
			ConfirmedDeposits: nil,
			BurnedDeposits:    nil,
			CommandBatches:    nil,
			Gateway:           Gateway{},
			Tokens:            nil,
		}
		chains = append(chains, chain)
	}
	return chains
}

// Validate validates the genesis state
func (m GenesisState) Validate() error {
	if !sort.SliceIsSorted(m.Chains, less(m.Chains)) {
		return getValidateError(0, fmt.Errorf("chains must be sorted by name (in params)"))
	}

	for j, chain := range m.Chains {
		if err := chain.Params.Validate(); err != nil {
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid params"))
		}

		if err := chain.Gateway.ValidateBasic(); err != nil {
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid gateway"))
		}

		if chain.Gateway.Status != GatewayStatusConfirmed {
			errStr := "gateway is not confirmed"

			if len(chain.Tokens) > 0 {
				return getValidateError(j, sdkerrors.Wrap(fmt.Errorf("cannot initialize tokens"), errStr))
			}

			if len(chain.ConfirmedDeposits) > 0 {
				return getValidateError(j, sdkerrors.Wrap(fmt.Errorf("cannot have confirmed deposits"), errStr))
			}

			if len(chain.BurnedDeposits) > 0 {
				return getValidateError(j, sdkerrors.Wrap(fmt.Errorf("cannot have burned deposits"), errStr))
			}

			if len(chain.BurnerInfos) > 0 {
				return getValidateError(j, sdkerrors.Wrap(fmt.Errorf("cannot have burned deposits"), errStr))
			}
		}

		for i, token := range chain.Tokens {
			if err := token.ValidateBasic(); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid token %d", i))
			}
		}

		for i, info := range chain.BurnerInfos {
			if err := info.ValidateBasic(); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid burner info %d", i))
			}

			if err := checkTokenInfo(info, chain.Tokens); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid burner info %d", i))
			}
		}

		for i, deposit := range chain.ConfirmedDeposits {
			if err := deposit.ValidateBasic(); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid confirmed deposit %d", i))
			}

			if err := checkBurnerInfo(deposit, chain.BurnerInfos); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid confirmed deposit %d", i))
			}
		}

		for i, deposit := range chain.BurnedDeposits {
			if err := deposit.ValidateBasic(); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid burned deposit %d", i))
			}

			if err := checkBurnerInfo(deposit, chain.BurnerInfos); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid burned deposit %d", i))
			}
		}

		if err := validateCommandBatches(chain.CommandBatches); err != nil {
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid command batches"))
		}

		if err := validateCommandBatches(chain.CommandBatches); err != nil {
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid command batches"))
		}

		queueState := make(map[string]codec.ProtoMarshaler, len(chain.CommandQueue))
		for key, value := range chain.CommandQueue {
			queueState[key] = &value
		}
		if err := utils.ValidateQueueState(queueState); err != nil {
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid command queue state"))
		}
	}

	return nil
}

func validateCommandBatches(batches []CommandBatchMetadata) error {
	var batchesWithoutPreviousBatch []string
	var batchesWithoutCompleteSign []string

	for i, batch := range batches {
		if batch.Status == BatchNonExistent {
			return fmt.Errorf("status of command batch %d not set", i)
		}

		if batch.Status != BatchSigned {
			batchesWithoutCompleteSign = append(batchesWithoutCompleteSign, strconv.Itoa(i))
		}

		if batch.ID == nil {
			return fmt.Errorf("ID of command batch %d not set", i)
		}

		if batch.PrevBatchedCommandsID == nil {
			batchesWithoutPreviousBatch = append(batchesWithoutPreviousBatch, strconv.Itoa(i))
		}
	}

	if len(batchesWithoutCompleteSign) > 1 {
		return fmt.Errorf("multiple uncompleted command batches: %s", strings.Join(batchesWithoutCompleteSign, ", "))
	}

	if len(batchesWithoutPreviousBatch) > 1 {
		return fmt.Errorf("multiple command batches without previous batch ID: %s", strings.Join(batchesWithoutPreviousBatch, ", "))
	}

	return nil
}

func checkBurnerInfo(deposit ERC20Deposit, burnerInfos []BurnerInfo) error {
	for _, info := range burnerInfos {
		if bytes.Equal(deposit.BurnerAddress.Bytes(), info.BurnerAddress.Bytes()) {
			if info.Asset != deposit.Asset {
				return fmt.Errorf("expected asset %s, got %s", info.Asset, deposit.Asset)
			}

			if info.DestinationChain != deposit.DestinationChain {
				return fmt.Errorf("expected destination address %s, got %s", info.DestinationChain, deposit.DestinationChain)
			}

			return nil
		}
	}

	return fmt.Errorf("burner info for address %s not found", deposit.BurnerAddress.Hex())
}

func checkTokenInfo(info BurnerInfo, tokens []ERC20TokenMetadata) error {
	for _, token := range tokens {
		if bytes.Equal(info.TokenAddress.Bytes(), token.TokenAddress.Bytes()) {
			if token.Asset != info.Asset {
				return fmt.Errorf("expected asset %s, got %s", token.Asset, info.Asset)
			}

			if token.Details.Symbol != info.Symbol {
				return fmt.Errorf("expected symbol %s, got %s", token.Details.Symbol, info.Symbol)
			}

			return nil
		}
	}

	return fmt.Errorf("token with address %s not found", info.TokenAddress.Hex())
}

// GetGenesisStateFromAppState returns x/evm GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

func getValidateError(chainIdx int, err error) error {
	return sdkerrors.Wrapf(sdkerrors.Wrapf(err, "invalid chain %d", chainIdx), "genesis state for module %s is invalid", ModuleName)
}
