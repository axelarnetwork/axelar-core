package config

import (
	"time"

	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ValdConfig contains all necessary vald configurations
type ValdConfig struct {
	tss.TssConfig   `mapstructure:",squash"`
	BroadcastConfig `mapstructure:",squash"`

	EVMConfig []evm.EVMConfig `mapstructure:"axelar_bridge_evm"`
}

// DefaultValdConfig returns a configurations populated with default values
func DefaultValdConfig() ValdConfig {
	return ValdConfig{
		EVMConfig:       evm.DefaultConfig(),
		TssConfig:       tss.DefaultConfig(),
		BroadcastConfig: DefaultBroadcastConfig(),
	}
}

// BroadcastConfig is the configuration for transaction broadcasting
type BroadcastConfig struct {
	MaxRetries int           `mapstructure:"max-retries"`
	MinTimeout time.Duration `mapstructure:"min-timeout"`
	FeeGranter string        `mapstructure:"fee_granter"`
}

// DefaultBroadcastConfig returns a configurations populated with default values
func DefaultBroadcastConfig() BroadcastConfig {
	return BroadcastConfig{
		MaxRetries: 10,
		MinTimeout: 5 * time.Second,
	}
}
