package app

import (
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	btcTypes "github.com/axelarnetwork/axelar-core/x/btc_bridge/types"
	tssdTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

type Config struct {
	btcTypes.BtcConfig          `mapstructure:"axelar_bridge_btc"`
	tssdTypes.TssdConfig        `mapstructure:",squash"`
	broadcastTypes.ClientConfig `mapstructure:",squash"`
}

func DefaultConfig() Config {
	return Config{
		BtcConfig:    btcTypes.DefaultConfig(),
		TssdConfig:   tssdTypes.TssdConfig{},
		ClientConfig: broadcastTypes.ClientConfig{},
	}
}
