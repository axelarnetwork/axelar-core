package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
			BurnerInfos:         nil,
			CommandQueue:        utils.QueueState{},
			ConfirmedDeposits:   nil,
			BurnedDeposits:      nil,
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
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid params"))
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

		if err := chain.CommandQueue.ValidateBasic(); err != nil {
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid command queue state"))
		}

		for _, event := range chain.Events {
			if eventSeen[string(event.GetID())] {
				return getValidateError(j, fmt.Errorf("duplicate event %s", event.GetID()))
			}

			if event.Status == EventNonExistent {
				return getValidateError(j, fmt.Errorf("invalid status of event %s", event.GetID()))
			}

			if err := event.ValidateBasic(); err != nil {
				return getValidateError(j, sdkerrors.Wrapf(err, "invalid event %s", event.GetID()))
			}

			eventSeen[string(event.GetID())] = true
		}

		if err := chain.ConfirmedEventQueue.ValidateBasic(); err != nil {
			return getValidateError(j, sdkerrors.Wrapf(err, "invalid confirmed event queue state"))
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
