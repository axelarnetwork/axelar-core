package config

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/config"
)

// Seed is used to unmarshal the seeds.toml file
type Seed struct {
	Address string `mapstructure:"address"`
}

// ReadSeeds read seeds from a seeds.toml file in the config directory
func ReadSeeds(v *viper.Viper) ([]string, error) {
	v.SetConfigName("seeds")
	v.SetConfigType("toml")

	err := v.MergeInConfig()
	if err != nil {
		return nil, err
	}

	var s []Seed
	if err := v.UnmarshalKey("seed", &s); err != nil {
		return nil, err
	}
	var seeds []string
	for _, seed := range s {
		seeds = append(seeds, seed.Address)
	}

	return seeds, nil
}

// MergeSeeds adds the additional seeds to the existing config and ensures they are all distinct
func MergeSeeds(cfg *config.Config, newSeeds []string) *config.Config {
	seeds := append(strings.Split(cfg.P2P.Seeds, ","), newSeeds...)
	cfg.P2P.Seeds = strings.Join(distinct(seeds), ",")

	return cfg
}

func distinct(slice []string) []string {
	seen := make(map[string]struct{})
	var distinctSlice []string
	for _, s := range slice {
		if _, ok := seen[s]; ok {
			continue
		}

		seen[s] = struct{}{}
		distinctSlice = append(distinctSlice, s)
	}
	return distinctSlice
}
