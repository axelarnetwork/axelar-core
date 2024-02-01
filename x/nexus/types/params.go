package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	params "github.com/cosmos/cosmos-sdk/x/params/types"

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
	// KeyGateway represents the key for the gateway's address
	KeyGateway = []byte("gateway")
	// KeyEndBlockerLimit represents the key for the end blocker limit
	KeyEndBlockerLimit = []byte("endBlockerLimit")
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
		Gateway:                               sdk.AccAddress{},
		EndBlockerLimit:                       50,
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
		params.NewParamSetPair(KeyGateway, &m.Gateway, validateGateway),
		params.NewParamSetPair(KeyEndBlockerLimit, &m.EndBlockerLimit, validateEndBlockerLimit),
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

	if err := validateGateway(m.Gateway); err != nil {
		return err
	}

	if err := validateEndBlockerLimit(m.EndBlockerLimit); err != nil {
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

func validateGateway(i interface{}) error {
	val, ok := i.(sdk.AccAddress)
	if !ok {
		return fmt.Errorf("invalid parameter type for Gateway: %T", i)
	}

	if len(val) == 0 {
		return nil
	}

	if err := sdk.VerifyAddressFormat(val); err != nil {
		return sdkerrors.Wrap(err, "invalid Gateway")
	}

	return nil
}

func validateEndBlockerLimit(limit interface{}) error {
	v, ok := limit.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type for end blocker limit: %T", limit)
	}
	if v == 0 {
		return fmt.Errorf("end blocker limit must be >0")
	}

	return nil
}
