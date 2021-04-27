package utils

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// OverwriteFlagDefaults overwrites the default values for already defined flags for the given command and all its children.
// Changes the currently set value if the flag is set.
func OverwriteFlagDefaults(c *cobra.Command, defaults map[string]string, updateVal bool) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			if updateVal {
				_ = c.Flags().Set(key, val)
				_ = c.PersistentFlags().Set(key, val)
			}
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		OverwriteFlagDefaults(c, defaults, updateVal)
	}
}
