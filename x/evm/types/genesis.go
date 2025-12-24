package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/axelarnetwork/axelar-core/utils"
)

// NewGenesisState returns a new genesis state
func NewGenesisState(chains []GenesisState_Chain) GenesisState {
	sort.Slice(chains, less(chains))

	for _, chain := range chains {
		sort.SliceStable(chain.Events, func(i, j int) bool {
			return chain.Events[i].Index < chain.Events[j].Index
		})
	}

	return GenesisState{Chains: chains}
}

func less(chains []GenesisState_Chain) func(i int, j int) bool {
	return func(i, j int) bool {
		return chains[i].Params.Chain.String() < chains[j].Params.Chain.String()
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultChains())
}

// DefaultChains returns the default chains for a genesis state
func DefaultChains() []GenesisState_Chain {
	var chains []GenesisState_Chain
	for _, params := range DefaultParams() {
		chain := GenesisState_Chain{
			Params:              params,
			CommandQueue:        utils.QueueState{},
			CommandBatches:      nil,
			Gateway:             Gateway{},
			Tokens:              nil,
			Events:              nil,
			ConfirmedEventQueue: utils.QueueState{},
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

	// events should be globally unique across all the chains
	eventSeen := make(map[string]bool)

	for j, chain := range m.Chains {
		if err := chain.Params.Validate(); err != nil {
			return getValidateError(j, errorsmod.Wrapf(err, "invalid params"))
		}

		if chain.Gateway.Address.IsZeroAddress() && len(chain.Tokens) > 0 {
			return getValidateError(j, errorsmod.Wrap(fmt.Errorf("cannot initialize tokens"), "gateway is not set"))
		}

		for i, token := range chain.Tokens {
			if err := token.ValidateBasic(); err != nil {
				return getValidateError(j, errorsmod.Wrapf(err, "invalid token %d", i))
			}
		}

		if err := validateCommandBatches(chain.CommandBatches); err != nil {
			return getValidateError(j, errorsmod.Wrapf(err, "invalid command batches"))
		}

		if err := chain.CommandQueue.ValidateBasic(); err != nil {
			return getValidateError(j, errorsmod.Wrapf(err, "invalid command queue state"))
		}

		for _, event := range chain.Events {
			if eventSeen[string(event.GetID())] {
				return getValidateError(j, fmt.Errorf("duplicate event %s", event.GetID()))
			}

			if event.Status == EventNonExistent {
				return getValidateError(j, fmt.Errorf("invalid status of event %s", event.GetID()))
			}

			if err := event.ValidateBasic(); err != nil {
				return getValidateError(j, errorsmod.Wrapf(err, "invalid event %s", event.GetID()))
			}

			eventSeen[string(event.GetID())] = true
		}

		if err := chain.ConfirmedEventQueue.ValidateBasic(); err != nil {
			return getValidateError(j, errorsmod.Wrapf(err, "invalid confirmed event queue state"))
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

		if i > 0 && !bytes.Equal(batches[i-1].ID, batch.PrevBatchedCommandsID) {
			return fmt.Errorf("previous batch ID mismatch at index %d (want '%s', got '%s')",
				i, hex.EncodeToString(batches[i-1].ID), hex.EncodeToString(batch.PrevBatchedCommandsID))
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
	return errorsmod.Wrapf(errorsmod.Wrapf(err, "invalid chain %d", chainIdx), "genesis state for module %s is invalid", ModuleName)
}
