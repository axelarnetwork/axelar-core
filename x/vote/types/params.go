package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

// Parameter store keys
var (
	KeyDefaultVotingThreshold = []byte("DefaultVotingThreshold")
	KeyEndBlockerLimit        = []byte("endBlockerLimit")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams - the module's default parameters
func DefaultParams() Params {
	return Params{
		DefaultVotingThreshold: utils.NewThreshold(2, 3),
		EndBlockerLimit:        100,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of vote module's parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDefaultVotingThreshold, &m.DefaultVotingThreshold, validateDefaultVotingThreshold),
		paramtypes.NewParamSetPair(KeyEndBlockerLimit, &m.EndBlockerLimit, validateEndBlockerLimit),
	}
}

// Validate performs a validation check on the parameters
func (m Params) Validate() error {
	if err := validateDefaultVotingThreshold(m.DefaultVotingThreshold); err != nil {
		return err
	}

	if err := validateEndBlockerLimit(m.EndBlockerLimit); err != nil {
		return err
	}

	return nil
}

func validateDefaultVotingThreshold(i interface{}) error {
	v, ok := i.(utils.Threshold)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LTE(utils.ZeroThreshold) || v.GT(utils.OneThreshold) {
		return fmt.Errorf("default voting threshold must be >0 and <=1: %s", v.String())
	}

	return nil
}

func validateEndBlockerLimit(limit interface{}) error {
	h, ok := limit.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type for end blocker limit: %T", limit)
	}
	if h <= 0 {
		return sdkerrors.Wrap(types.ErrInvalidGenesis, "end blocker limit must be greater >0")
	}

	return nil
}
