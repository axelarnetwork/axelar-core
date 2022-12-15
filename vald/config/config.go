package config

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// ValdConfig contains all necessary vald configurations
type ValdConfig struct {
	tss.TssConfig                `mapstructure:"tss"`
	BroadcastConfig              `mapstructure:"broadcast"`
	BatchSizeLimit               int           `mapstructure:"max_batch_size"`
	BatchThreshold               int           `mapstructure:"batch_threshold"`
	MaxBlocksBehindLatest        int64         `mapstructure:"max_blocks_behind_latest"` // The max amount of blocks behind the latest until which the cached height is considered valid.
	EventNotificationsMaxRetries int           `mapstructure:"event_notifications_max_retries"`
	EventNotificationsBackOff    time.Duration `mapstructure:"event_notifications_back_off"`
	MaxLatestBlockAge            time.Duration `mapstructure:"max_latest_block_age"`  // If a block is older than this, vald does not consider it to be the latest block. This is supposed to be sufficiently larger than the block production time.
	NoNewBlockPanicTimeout       time.Duration `mapstructure:"no_new_blocks_timeout"` // At times vald stalls completely. Until the bug is found it is better to panic and allow users to restart the process instead of doing nothing. Once at least one block has been seen vald will panic if it does not see another before the timout expires.

	EVMConfig []evm.EVMConfig `mapstructure:"axelar_bridge_evm"`
}

// DefaultValdConfig returns a configurations populated with default values
func DefaultValdConfig() ValdConfig {
	return ValdConfig{
		TssConfig:                    tss.DefaultConfig(),
		BroadcastConfig:              DefaultBroadcastConfig(),
		BatchSizeLimit:               250,
		BatchThreshold:               3,
		MaxBlocksBehindLatest:        10, // Max voting/sign/heartbeats periods are under 10 blocks
		MaxLatestBlockAge:            15 * time.Second,
		EVMConfig:                    evm.DefaultConfig(),
		EventNotificationsMaxRetries: 3,
		EventNotificationsBackOff:    1 * time.Second,
		NoNewBlockPanicTimeout:       2 * time.Minute,
	}
}

// BroadcastConfig is the configuration for transaction broadcasting
type BroadcastConfig struct {
	MaxRetries          int            `mapstructure:"max_retries"`
	MinSleepBeforeRetry time.Duration  `mapstructure:"min_sleep_before_retry"`
	MaxTimeout          time.Duration  `mapstructure:"max_timeout"`
	FeeGranter          sdk.AccAddress `mapstructure:"fee_granter"`
}

// DefaultBroadcastConfig returns a configurations populated with default values
func DefaultBroadcastConfig() BroadcastConfig {
	return BroadcastConfig{
		MaxRetries:          3,
		MinSleepBeforeRetry: 5 * time.Second,
		MaxTimeout:          15 * time.Second,
	}
}
