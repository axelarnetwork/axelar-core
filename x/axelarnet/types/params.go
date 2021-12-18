package types

import (
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParamspace - default parameter namespace
const (
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	KeyAssets             = []byte("assetInfo")
	KeyRouteTimeoutWindow = []byte("routeTimeoutWindow")
	KeyTransactionFeeRate = []byte("transactionFeeRate")
	KeyMinAmount          = []byte("minAmount")
)

// KeyTable retrieves a subspace table for the module
func KeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams creates the default genesis parameters
func DefaultParams() Params {
	return Params{
		RouteTimeoutWindow: 17000,
		TransactionFeeRate: sdktypes.NewDecWithPrec(1, 3), // 0.1%
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
		params.NewParamSetPair(KeyTransactionFeeRate, &m.TransactionFeeRate, validateTransactionFeeRate),
	}
}

// Validate checks if the parameters are valid
func (m Params) Validate() error {
	if err := validatePosUInt64("RouteTimeoutWindow")(m.RouteTimeoutWindow); err != nil {
		return err
	}

	if err := validateTransactionFeeRate(m.TransactionFeeRate); err != nil {
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

		if val <= 0 {
			return fmt.Errorf("%s must be a positive integer", field)
		}

		return nil
	}
}

func validateTransactionFeeRate(i interface{}) error {
	v, ok := i.(sdktypes.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("transaction fee rate must be positive: %s", v)
	}

	if v.GT(sdktypes.OneDec()) {
		return fmt.Errorf("transaction fee rate %s must be <= 1", v)
	}

	return nil
}
