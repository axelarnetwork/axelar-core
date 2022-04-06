package config

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ValdConfig contains all necessary vald configurations
type ValdConfig struct {
	tss.TssConfig      `mapstructure:",squash"`
	BroadcastConfig    `mapstructure:",squash"`
	BatchSizeLimit     int   `mapstructure:"max_batch_size"`
	BatchThreshold     int   `mapstructure:"batch_threshold"`
	MaxOutOfSyncHeight int64 `mapstructure:"max_out_of_sync_height"`

	EVMConfig []evm.EVMConfig `mapstructure:"axelar_bridge_evm"`
}

// DefaultValdConfig returns a configurations populated with default values
func DefaultValdConfig() ValdConfig {
	return ValdConfig{
		TssConfig:          tss.DefaultConfig(),
		BroadcastConfig:    DefaultBroadcastConfig(),
		BatchSizeLimit:     250,
		BatchThreshold:     3,
		MaxOutOfSyncHeight: 50, // Max voting/sign/heartbeats periods are 50 blocks
		EVMConfig:          evm.DefaultConfig(),
	}
}

// BroadcastConfig is the configuration for transaction broadcasting
type BroadcastConfig struct {
	MaxRetries int            `mapstructure:"max_retries"`
	MinTimeout time.Duration  `mapstructure:"min_timeout"`
	FeeGranter sdk.AccAddress `mapstructure:"fee_granter"`
}

// DefaultBroadcastConfig returns a configurations populated with default values
func DefaultBroadcastConfig() BroadcastConfig {
	return BroadcastConfig{
		MaxRetries: 10,
		MinTimeout: 5 * time.Second,
	}
}
