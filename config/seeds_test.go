package config_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	tmcfg "github.com/tendermint/tendermint/config"

	"github.com/axelarnetwork/axelar-core/config"
)

func TestReadSeeds(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		v := viper.New()
		v.AddConfigPath("./test_files")

		seeds, err := config.ReadSeeds(v)
		assert.NoError(t, err)

		assert.Len(t, seeds, 16)
		for _, seed := range seeds {
			assert.NotEmpty(t, seed)
		}
	})

	t.Run("file exists but is empty", func(t *testing.T) {
		v := viper.New()
		v.AddConfigPath("./empty_test_files")

		seeds, err := config.ReadSeeds(v)
		assert.NoError(t, err)
		assert.Len(t, seeds, 0)
	})

	t.Run("file does not exist", func(t *testing.T) {
		v := viper.New()
		v.AddConfigPath("some other path")

		seeds, err := config.ReadSeeds(v)
		assert.ErrorAs(t, err, &viper.ConfigFileNotFoundError{})

		assert.Len(t, seeds, 0)
	})
}

func TestMergeSeeds(t *testing.T) {
	cfg := &tmcfg.Config{P2P: &tmcfg.P2PConfig{Seeds: "a,b,c,d"}}
	cfg = config.MergeSeeds(cfg, []string{"b", "d", "f", "z"})

	assert.Equal(t, "a,b,c,d,f,z", cfg.P2P.Seeds)
}
