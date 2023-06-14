package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	KeyRouteTimeoutWindow = []byte("routeTimeoutWindow")
	KeyTransferLimit      = []byte("transferLimit")
	KeyEndBlockerLimit    = []byte("endBlockerLimit")
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
		RouteTimeoutWindow:               17000,
		TransferLimit:                    20,
		EndBlockerLimit:                  50,
		CallContractsProposalMinDeposits: nil,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of tss module's parameters.
func (m *Params) ParamSetPairs() params.ParamSetPairs {
	/*
		because the subspace package makes liberal use of pointers to set and get values from the store,
		this method needs to have a pointer receiver AND NewParamSetPair needs to receive the
		parameter values as pointer arguments, otherwise either the internal type reflection panics or the value will not be
		set on the correct Params data struct
	*/
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyRouteTimeoutWindow, &m.RouteTimeoutWindow, validatePosUInt64("RouteTimeoutWindow")),
		params.NewParamSetPair(KeyTransferLimit, &m.TransferLimit, validatePosUInt64("TransferLimit")),
		params.NewParamSetPair(KeyEndBlockerLimit, &m.EndBlockerLimit, validatePosUInt64("EndBlockerLimit")),
		params.NewParamSetPair(KeyCallContractsProposalMinDeposits, &m.CallContractsProposalMinDeposits, validateCallContractsProposalMinDeposits),
	}
}

// Validate checks if the parameters are valid
func (m Params) Validate() error {
	if err := validatePosUInt64("RouteTimeoutWindow")(m.RouteTimeoutWindow); err != nil {
		return err
	}

	if err := validatePosUInt64("TransferLimit")(m.TransferLimit); err != nil {
		return err
	}

	if err := validatePosUInt64("EndBlockerLimit")(m.EndBlockerLimit); err != nil {
		return err
	}

	if err := validateCallContractsProposalMinDeposits(m.CallContractsProposalMinDeposits); err != nil {
		return err
	}

	return nil
}

func validatePosUInt64(field string) func(value interface{}) error {
	return func(value interface{}) error {
		val, ok := value.(uint64)
		if !ok {
			return fmt.Errorf("invalid parameter type for %s: %T", field, value)
		}

		if val == 0 {
			return fmt.Errorf("%s must be a positive integer", field)
		}

		return nil
	}
}

func validateCallContractsProposalMinDeposits(i interface{}) error {
	val, ok := i.(CallContractProposalMinDeposits)
	if !ok {
		return fmt.Errorf("invalid parameter type for CallContractsProposalMinDeposits: %T", i)
	}

	if err := val.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "invalid CallContractsProposalMinDeposits")
	}

	return nil
}
