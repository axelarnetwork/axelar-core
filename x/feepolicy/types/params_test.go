package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/feepolicy/types"
)

func TestParamsValidate(t *testing.T) {
	t.Run("default params are valid and allow only uaxl", func(t *testing.T) {
		params := types.DefaultParams()
		assert.NoError(t, params.Validate())
		assert.Equal(t, []string{"uaxl"}, params.AllowedFeeDenoms)
	})

	t.Run("multiple valid denoms are accepted", func(t *testing.T) {
		params := types.Params{AllowedFeeDenoms: []string{"uaxl", "ibc/ABCD"}}
		assert.NoError(t, params.Validate())
	})

	t.Run("empty allowlist is rejected", func(t *testing.T) {
		params := types.Params{AllowedFeeDenoms: nil}
		assert.Error(t, params.Validate())
	})

	t.Run("invalid denom is rejected", func(t *testing.T) {
		params := types.Params{AllowedFeeDenoms: []string{"!bad!"}}
		assert.Error(t, params.Validate())
	})

	t.Run("duplicate denom is rejected", func(t *testing.T) {
		params := types.Params{AllowedFeeDenoms: []string{"uaxl", "uaxl"}}
		assert.Error(t, params.Validate())
	})
}

func TestDefaultGenesisState(t *testing.T) {
	assert.NoError(t, types.DefaultGenesisState().Validate())
}
