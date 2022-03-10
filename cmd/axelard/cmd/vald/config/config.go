package config

import (
	"time"

	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValdConfig contains all necessary vald configurations
type ValdConfig struct {
	tss.TssConfig   `mapstructure:",squash"`
	BroadcastConfig `mapstructure:",squash"`
	BatchSizeLimit  int `mapstructure:"max_batch_size"`
	BatchThreshold  int `mapstructure:"batch_threshold"`

	EVMConfig []evm.EVMConfig `mapstructure:"axelar_bridge_evm"`
}

// DefaultValdConfig returns a configurations populated with default values
func DefaultValdConfig() ValdConfig {
	return ValdConfig{
		TssConfig:       tss.DefaultConfig(),
		BroadcastConfig: DefaultBroadcastConfig(),
		BatchSizeLimit:  250,
		BatchThreshold:  3,
		EVMConfig:       evm.DefaultConfig(),
	}
}

// BroadcastConfig is the configuration for transaction broadcasting
type BroadcastConfig struct {
	MaxRetries int            `mapstructure:"max-retries"`
	MinTimeout time.Duration  `mapstructure:"min-timeout"`
	FeeGranter sdk.AccAddress `mapstructure:"fee_granter"`
}

// DefaultBroadcastConfig returns a configurations populated with default values
func DefaultBroadcastConfig() BroadcastConfig {
	return BroadcastConfig{
		MaxRetries: 10,
		MinTimeout: 5 * time.Second,
	}
}
