package types

import (
	"time"
)

type BtcConfig struct {
	RpcAddr        string        `mapstructure:"rpc_addr"`
	CookiePath     string        `mapstructure:"cookie_file"`
	RPCTimeout     time.Duration `mapstructure:"timeout_rpc"`
	StartUpTimeout time.Duration `mapstructure:"timeout_startup"`
}
