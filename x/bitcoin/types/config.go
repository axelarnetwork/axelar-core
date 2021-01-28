package types

import (
	"time"
)

type BtcConfig struct {
	RPCAddr        string        `mapstructure:"rpc_addr"`
	RPCUser        string        `mapstructure:"rpc_user"`
	RPCPass        string        `mapstructure:"rpc_pass"`
	CookiePath     string        `mapstructure:"cookie_file"`
	RPCTimeout     time.Duration `mapstructure:"timeout_rpc"`
	StartUpTimeout time.Duration `mapstructure:"timeout_startup"`
	WithBtcBridge  bool          `mapstructure:"start-with-bridge"`
}

func DefaultConfig() BtcConfig {
	return BtcConfig{
		RPCAddr:        "localhost:8332",
		RPCTimeout:     60 * time.Second,
		StartUpTimeout: 100 * time.Second,
		WithBtcBridge:  true,
	}
}
