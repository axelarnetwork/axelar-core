package cmd

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	root, _ := NewRootCmd()
	t.Run("broadcast flag correctly set", func(t *testing.T) {
		cmd := root
		assert.True(t, testCmd(t, cmd), "no command with usage 'vald-start' found")
	})
}

func testCmd(t *testing.T, cmd *cobra.Command) (foundVald bool) {
	if f := cmd.Flags().Lookup(flags.FlagBroadcastMode); f != nil {
		if cmd.Use == "vald-start" {
			foundVald = true
			assert.Equal(t, flags.BroadcastSync, f.Value.String())
		} else {
			assert.Equal(t, flags.BroadcastBlock, f.Value.String())
		}
	}

	for _, c := range cmd.Commands() {
		foundVald = foundVald || testCmd(t, c)
	}
	return foundVald
}
