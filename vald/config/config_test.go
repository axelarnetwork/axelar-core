package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
)

func TestConfigAlias(t *testing.T) {
	fp, err := buildTestdataFilePath()
	assert.NoError(t, err)

	v := viper.GetViper()
	v.AddConfigPath(fp)
	v.SetConfigName("config.toml")
	v.SetConfigType("toml")
	assert.NoError(t, v.ReadInConfig())

	v.RegisterAlias("broadcast.max_timeout", "rpc.timeout_broadcast_tx_commit")

	var conf ValdConfig
	assert.NoError(t, v.Unmarshal(&conf, AddDecodeHooks))

	assert.Equal(t, 99*time.Hour, conf.MaxTimeout)
	assert.Equal(t, 1*time.Nanosecond, conf.MinSleepBeforeRetry)
	assert.Len(t, conf.EVMConfig, 2)
	assert.Equal(t, rpc.Confirmation, conf.EVMConfig[0].FinalityOverride)
	assert.Equal(t, rpc.NoOverride, conf.EVMConfig[1].FinalityOverride)
}

func buildTestdataFilePath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fp := filepath.Join(wd, "testdata")
	return fp, nil
}
