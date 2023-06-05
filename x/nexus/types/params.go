package types

import (
	"fmt"
	"sort"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
)

var (

	// KeyChainActivationThreshold represents the key for chain activation threshold
	KeyChainActivationThreshold = []byte("chainActivationThreshold")
	// KeyChainMaintainerMissingVoteThreshold represents the key for chain maintainer missing vote threshold
	KeyChainMaintainerMissingVoteThreshold = []byte("chainMaintainerMissingVoteThreshold")
	// KeyChainMaintainerIncorrectVoteThreshold represents the key for chain maintainer incorrect vote threshold
	KeyChainMaintainerIncorrectVoteThreshold = []byte("chainMaintainerIncorrectVoteThreshold")
	// KeyChainMaintainerCheckWindow represents the key for chain maintainer check window
	KeyChainMaintainerCheckWindow = []byte("chainMaintainerCheckWindow")
	// KeyCallContractsProposalMinDeposits represents the key for call contracts proposal min deposits
	KeyCallContractsProposalMinDeposits = []byte("callContractsProposalMinDeposits")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{
		ChainActivationThreshold:              utils.NewThreshold(55, 100),
		ChainMaintainerMissingVoteThreshold:   utils.NewThreshold(20, 100),
		ChainMaintainerIncorrectVoteThreshold: utils.NewThreshold(15, 100),
		ChainMaintainerCheckWindow:            500,
		CallContractsProposalMinDeposits:      make(map[string]Params_Coins),
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of nexus module's parameters.
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyChainActivationThreshold, &m.ChainActivationThreshold, validateThresholdWith("ChainActivationThreshold")),
		params.NewParamSetPair(KeyChainMaintainerMissingVoteThreshold, &m.ChainMaintainerMissingVoteThreshold, validateThresholdWith("ChainMaintainerMissingVoteThreshold")),
		params.NewParamSetPair(KeyChainMaintainerIncorrectVoteThreshold, &m.ChainMaintainerIncorrectVoteThreshold, validateThresholdWith("ChainMaintainerIncorrectVoteThreshold")),
		params.NewParamSetPair(KeyChainMaintainerCheckWindow, &m.ChainMaintainerCheckWindow, validateChainMaintainerCheckWindow),
		params.NewParamSetPair(KeyCallContractsProposalMinDeposits, &m.CallContractsProposalMinDeposits, validateCallContractsProposalMinDeposits),
	}
}

// Validate checks if the parameters are valid
func (m Params) Validate() error {
	if err := validateThresholdWith("ChainActivationThreshold")(m.ChainActivationThreshold); err != nil {
		return err
	}

	if err := validateThresholdWith("ChainMaintainerMissingVoteThreshold")(m.ChainMaintainerMissingVoteThreshold); err != nil {
		return err
	}

	if err := validateThresholdWith("ChainMaintainerIncorrectVoteThreshold")(m.ChainMaintainerIncorrectVoteThreshold); err != nil {
		return err
	}

	if err := validateChainMaintainerCheckWindow(m.ChainMaintainerCheckWindow); err != nil {
		return err
	}

	if err := validateCallContractsProposalMinDeposits(m.CallContractsProposalMinDeposits); err != nil {
		return err
	}

	return nil
}

func validateThresholdWith(paramName string) func(interface{}) error {
	return func(i interface{}) error {
		val, ok := i.(utils.Threshold)
		if !ok {
			return fmt.Errorf("invalid parameter type for %s: %T", paramName, i)
		}

		if err := val.Validate(); err != nil {
			return sdkerrors.Wrapf(err, "invalid %s", paramName)
		}

		return nil
	}
}

func validateChainMaintainerCheckWindow(i interface{}) error {
	val, ok := i.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type for ChainMaintainerCheckWindow: %T", i)
	}

	if val <= 0 {
		return fmt.Errorf("ChainMaintainerCheckWindow must be >0")
	}

	if val >= maxBitmapSize {
		return fmt.Errorf("ChainMaintainerCheckWindow must be < %d", maxBitmapSize)
	}

	return nil
}

func validateCallContractsProposalMinDeposits(i interface{}) error {
	val, ok := i.(map[string]Params_Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type for CallContractsProposalMinDeposits: %T", i)
	}

	contractAddresses := maps.Keys(val)
	sort.Strings(contractAddresses)

	for _, contractAddress := range contractAddresses {
		if strings.ToLower(contractAddress) != contractAddress {
			return fmt.Errorf("contract addresses in CallContractsProposalMinDeposits must be lowercase")
		}

		if err := val[contractAddress].Coins.Validate(); err != nil {
			return err
		}
	}

	return nil
}
