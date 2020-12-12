package app

import (
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

type Config struct {
	bitcoin.BtcConfig      `mapstructure:"axelar_bridge_btc"`
	tss.TssConfig          `mapstructure:",squash"`
	broadcast.ClientConfig `mapstructure:",squash"`
}

func DefaultConfig() Config {
	return Config{
		BtcConfig:    bitcoin.DefaultConfig(),
		TssConfig:    tss.TssConfig{},
		ClientConfig: broadcast.ClientConfig{},
	}
}
