package types

import (
	"time"
)

// BtcConfig - configuration for bitcoin client
type BtcConfig struct {
	RPCAddr        string        `mapstructure:"rpc_addr"`
	RPCUser        string        `mapstructure:"rpc_user"`
	RPCPass        string        `mapstructure:"rpc_pass"`
	CookiePath     string        `mapstructure:"cookie_file"`
	RPCTimeout     time.Duration `mapstructure:"timeout_rpc"`
	StartUpTimeout time.Duration `mapstructure:"timeout_startup"`
}

// DefaultConfig returns a BtcConfig with default values
func DefaultConfig() BtcConfig {
	return BtcConfig{
		RPCAddr:        "localhost:8332",
		RPCTimeout:     60 * time.Second,
		StartUpTimeout: 100 * time.Second,
	}
}
