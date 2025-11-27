package config

import (
	"bytes"
	"github.com/axelarnetwork/axelar-core/app"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
)

// TestConfig_GoldenFile ensures that default config values don't get changed accidentally
// and that new config parameters are not forgotten. If this test fails, review the
// changes carefully and run with UPDATE_GOLDEN=true to update the golden file.
func TestConfig_GoldenFile(t *testing.T) {
	app.SetConfig()
	cfg := testConfig()

	var buf bytes.Buffer
	require.NoError(t, WriteTOML(&buf, cfg))
	serialized := buf.Bytes()

	goldenPath := filepath.Join(testdataPath(t), "golden_config.toml")

	if os.Getenv("UPDATE_GOLDEN") == "true" {
		require.NoError(t, os.WriteFile(goldenPath, serialized, 0644))
		t.Skip("Golden file updated")
	}

	golden, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "Golden file not found. Run with UPDATE_GOLDEN=true to create it.")

	assert.Equal(t, string(golden), string(serialized), "Config serialization changed. Run with UPDATE_GOLDEN=true to update.")
}

// TestConfig_RoundTrip ensures the toml can be decoded back into the expected values
func TestConfig_RoundTrip(t *testing.T) {
	goldenPath := filepath.Join(testdataPath(t), "golden_config.toml")

	v := viper.New()
	v.SetConfigFile(goldenPath)
	v.SetConfigType("toml")
	require.NoError(t, v.ReadInConfig())

	var loaded ValdConfig
	require.NoError(t, v.Unmarshal(&loaded, AddDecodeHooks))

	expected := testConfig()
	assert.Equal(t, expected, loaded)
}

func testConfig() ValdConfig {
	cfg := DefaultValdConfig()
	cfg.FeeGranter = sdk.AccAddress("test-fee-granter-addr")
	cfg.EVMConfig = []evmtypes.EVMConfig{
		{
			Name:             "ethereum",
			RPCAddr:          "https://eth.example.com",
			WithBridge:       true,
			FinalityOverride: rpc.Confirmation,
		},
		{
			Name:             "avalanche",
			RPCAddr:          "https://avax.example.com",
			WithBridge:       true,
			FinalityOverride: rpc.NoOverride,
		},
		{
			Name:             "polygon",
			RPCAddr:          "https://polygon.example.com",
			WithBridge:       false,
			FinalityOverride: rpc.NoOverride,
		},
	}
	return cfg
}

func testdataPath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "testdata")
}
