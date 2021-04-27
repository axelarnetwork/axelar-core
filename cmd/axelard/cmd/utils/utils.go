package utils

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// OverwriteFlagDefaults overwrites the default values for already defined flags
func OverwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		OverwriteFlagDefaults(c, defaults)
	}
}
