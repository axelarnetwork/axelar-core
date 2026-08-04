package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultGenesisState(t *testing.T) {
	assert.NoError(t, DefaultGenesisState().Validate())
}

func TestGenesisState_Validate(t *testing.T) {
	assert.NoError(t, NewGenesisState(Params{AllowedFeeDenoms: []string{"uaxl", "uusdc"}}).Validate())

	// an empty allowlist is invalid: it would leave no denom able to pay fees
	assert.Error(t, NewGenesisState(Params{AllowedFeeDenoms: nil}).Validate())

	// duplicate and malformed denoms are rejected
	assert.Error(t, NewGenesisState(Params{AllowedFeeDenoms: []string{"uaxl", "uaxl"}}).Validate())
	assert.Error(t, NewGenesisState(Params{AllowedFeeDenoms: []string{"UAXL!"}}).Validate())
}
