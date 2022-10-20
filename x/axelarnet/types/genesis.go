package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewGenesisState returns a new GenesisState instance
func NewGenesisState(p Params, feeCollector sdk.AccAddress, chains []CosmosChain, transferQueue utils.QueueState, transfers []IBCTransfer, seqIDMapping map[string]uint64) *GenesisState {
	SortChains(chains)
	return &GenesisState{
		Params:           p,
		CollectorAddress: feeCollector,
		Chains:           chains,
		TransferQueue:    transferQueue,
		IBCTransfers:     transfers,
		SeqIDMapping:     seqIDMapping,
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
		IBCTransfers:  nil,
		SeqIDMapping:  nil,
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
		if err := chain.ValidateBasic(); err != nil {
			return getValidateError(sdkerrors.Wrap(err, fmt.Sprintf("faulty chain entry %d", i)))
		}
	}

	if err := m.TransferQueue.ValidateBasic(); err != nil {
		return getValidateError(sdkerrors.Wrapf(err, "invalid transfer queue state"))
	}

	// IBCTransfer ID should be unique
	transferSeen := make(map[nexus.TransferID]bool)
	for _, t := range m.IBCTransfers {
		if transferSeen[t.ID] {
			return getValidateError(fmt.Errorf("duplicate transfer ID %d", t.ID))
		}

		if t.Status == TransferNonExistent {
			return getValidateError(fmt.Errorf("invalid status of transfer %s", t.ID))
		}

		if err := t.ValidateBasic(); err != nil {
			return getValidateError(sdkerrors.Wrapf(err, "invalid transfer %s", t.ID))
		}

		transferSeen[t.ID] = true
	}

	// IBCTransfer ID should be uniquely mapped
	transferIDSeen := make(map[uint64]bool)
	sortedKeys := SortedMapKeys(m.SeqIDMapping, strings.Compare)
	for _, seqKey := range sortedKeys {
		if transferIDSeen[m.SeqIDMapping[seqKey]] {
			return getValidateError(fmt.Errorf("duplicate transfer ID %d for %s", m.SeqIDMapping[seqKey], seqKey))
		}
	}

	return nil
}

func getValidateError(err error) error {
	return sdkerrors.Wrapf(err, "genesis state for module %s is invalid", ModuleName)
}

// SortedMapKeys returns sorted map keys
func SortedMapKeys[T comparable, S any](m map[T]S, compare func(t1, t2 T) int) []T {
	keys := maps.Keys(m)
	sort.SliceStable(keys, func(i, j int) bool { return compare(keys[i], keys[j]) < 0 })

	return keys
}
