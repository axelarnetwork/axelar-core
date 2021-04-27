package utils

import (
	"github.com/spf13/cobra"
)

// OverwriteFlagValues overwrites the values for already defined flags
func OverwriteFlagValues(c *cobra.Command, values map[string]string) {
	for key, val := range values {
		_ = c.Flags().Set(key, val)
		_ = c.PersistentFlags().Set(key, val)
	}
	for _, c := range c.Commands() {
		OverwriteFlagValues(c, values)
	}
}
