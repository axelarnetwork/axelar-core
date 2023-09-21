package cmd

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	root, _ := NewRootCmd()
	t.Run("broadcast flag correctly set", func(t *testing.T) {
		cmd := root
		assert.True(t, testCmd(t, cmd, flags.BroadcastSync), "no command with usage 'vald-start' found")
	})

	t.Run("keyring default set to file", func(t *testing.T) {
		cmd := root
		assert.True(t, keyringBackendSetToFile(t, cmd), "no keyring backend flag found")
	})
}

func keyringBackendSetToFile(t *testing.T, cmd *cobra.Command) (foundKeyringBackend bool) {
	if f := cmd.Flags().Lookup(flags.FlagKeyringBackend); f != nil {
		assert.Equal(t, keyring.BackendFile, f.DefValue)
		assert.Equal(t, keyring.BackendFile, f.Value.String())
		assert.False(t, f.Changed)
		foundKeyringBackend = true
	}

	for _, c := range cmd.Commands() {
		foundKeyringBackend = foundKeyringBackend || keyringBackendSetToFile(t, c)
	}
	return foundKeyringBackend
}

func testCmd(t *testing.T, cmd *cobra.Command, expectedMode string) (foundVald bool) {
	if f := cmd.Flags().Lookup(flags.FlagBroadcastMode); f != nil {
		if cmd.Use == "vald-start" {
			foundVald = true
			assert.Equal(t, expectedMode, f.Value.String())
			assert.False(t, f.Changed)
		}
	}

	for _, c := range cmd.Commands() {
		foundVald = foundVald || testCmd(t, c, expectedMode)
	}
	return foundVald
}
