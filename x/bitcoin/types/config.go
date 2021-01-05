package types

import (
	"os"
	"time"
)

type BtcConfig struct {
	RpcAddr        string        `mapstructure:"rpc_addr"`
	RpcUser        string        `mapstructure:"rpc_user"`
	RpcPass        string        `mapstructure:"rpc_pass"`
	CookiePath     string        `mapstructure:"cookie_file"`
	RPCTimeout     time.Duration `mapstructure:"timeout_rpc"`
	StartUpTimeout time.Duration `mapstructure:"timeout_startup"`
	WithBtcBridge  bool          `mapstructure:"start-with-bridge"`
}

func DefaultConfig() BtcConfig {
	return BtcConfig{
		RpcAddr:        "localhost:8332",
		CookiePath:     os.ExpandEnv("$HOME/.bitcoin/.cookie"),
		RPCTimeout:     60 * time.Second,
		StartUpTimeout: 100 * time.Second,
		WithBtcBridge:  true,
	}
}
