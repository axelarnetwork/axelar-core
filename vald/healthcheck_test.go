package vald

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/axelarnetwork/axelar-core/vald/evm/rpc"
)

func TestLoadValdConfigUsesDecodeHooks(t *testing.T) {
	v := viper.New()
	v.Set("tss.tofnd-host", "configured-host")
	v.Set("tss.tofnd-port", "50052")
	v.Set("tss.tofnd-dial-timeout", "7s")
	v.Set("axelar_bridge_evm", []map[string]any{
		{
			"name":              "linea",
			"finality_override": "confirmation",
		},
	})

	cfg, err := loadValdConfig(&server.Context{Viper: v})
	require.NoError(t, err)
	require.Equal(t, "configured-host", cfg.TssConfig.Host)
	require.Equal(t, "50052", cfg.TssConfig.Port)
	require.Equal(t, 7*time.Second, cfg.TssConfig.DialTimeout)
	require.Len(t, cfg.EVMConfig, 1)
	require.Equal(t, rpc.Confirmation, cfg.EVMConfig[0].FinalityOverride)
}
