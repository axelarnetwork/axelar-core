package app

import (
	vald "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/config"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// Config contains all necessary application configurations
type Config struct {
	bitcoin.BtcConfig `mapstructure:"axelar_bridge_btc"`
	tss.TssConfig     `mapstructure:",squash"`
	vald.ClientConfig `mapstructure:",squash"`

	EVMConfig []evm.EVMConfig `mapstructure:"axelar_bridge_evm"`
}

// DefaultConfig returns a configurations populated with default values
func DefaultConfig() Config {
	return Config{
		EVMConfig:    evm.DefaultConfig(),
		BtcConfig:    bitcoin.DefaultConfig(),
		TssConfig:    tss.TssConfig{},
		ClientConfig: vald.ClientConfig{},
	}
}
