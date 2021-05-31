package app

import (
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	broadcast "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// Config contains all necessary application configurations
type Config struct {
	evm.EthConfig          `mapstructure:"axelar_bridge_eth"`
	bitcoin.BtcConfig      `mapstructure:"axelar_bridge_btc"`
	tss.TssConfig          `mapstructure:",squash"`
	broadcast.ClientConfig `mapstructure:",squash"`
}

// DefaultConfig returns a configurations populated with default values
func DefaultConfig() Config {
	return Config{
		EthConfig:    evm.DefaultConfig(),
		BtcConfig:    bitcoin.DefaultConfig(),
		TssConfig:    tss.TssConfig{},
		ClientConfig: broadcast.ClientConfig{},
	}
}
