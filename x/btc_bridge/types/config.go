package types

import (
	"os"
	"time"
)

type BtcConfig struct {
	RpcAddr            string        `mapstructure:"rpc_addr"`
	CookiePath         string        `mapstructure:"cookie_file"`
	RPCTimeout         time.Duration `mapstructure:"timeout_rpc"`
	StartUpTimeout     time.Duration `mapstructure:"timeout_startup"`
	WithBtcBridge      bool          `mapstructure:"start-with-bridge"`
	ConfirmationHeight int64         `mapstructure:"confirmation-height"`
}

func DefaultConfig() BtcConfig {
	return BtcConfig{
		RpcAddr:            "localhost:8332",
		CookiePath:         os.ExpandEnv("$HOME/.bitcoin/.cookie"),
		RPCTimeout:         60 * time.Second,
		StartUpTimeout:     100 * time.Second,
		WithBtcBridge:      true,
		ConfirmationHeight: 6,
	}
}
